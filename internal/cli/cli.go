package cli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/harryvince/danger-go/internal/config"
	"github.com/harryvince/danger-go/internal/danger"
	"github.com/harryvince/danger-go/internal/git"
	"github.com/harryvince/danger-go/internal/github"
	"github.com/harryvince/danger-go/internal/version"
	"github.com/harryvince/danger-go/schema"
)

const usage = `danger-go

Usage:
  danger-go local [--config path]
  danger-go ci    [--config path]
  danger-go validate [--config path]
  danger-go version

Commands:
  local   Run checks against the local git working tree.
  ci      Run checks using GitHub Actions metadata when available.
  validate Validate a danger-go config file against the JSON Schema.
  version Print the danger-go version.
`

func Run(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		fmt.Fprint(stderr, usage)
		return nil
	}

	switch args[0] {
	case "local", "ci":
		return runChecks(ctx, args[0], args[1:], stdout)
	case "validate":
		return runValidate(args[1:], stdout)
	case "version":
		fmt.Fprintln(stdout, version.Info())
		return nil
	case "-h", "--help", "help":
		fmt.Fprint(stdout, usage)
		return nil
	default:
		return fmt.Errorf("unknown command %q\n\n%s", args[0], usage)
	}
}

func runValidate(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("validate", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	configPath := fs.String("config", "", "path to .danger.yaml or .danger.yml")
	if err := fs.Parse(args); err != nil {
		return err
	}

	path := *configPath
	if path == "" {
		var err error
		path, err = config.FindPath()
		if err != nil {
			return err
		}
	}

	if err := schema.ValidateFile(path); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "Config is valid: %s\n", path)
	return nil
}

func runChecks(ctx context.Context, mode string, args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet(mode, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	configPath := fs.String("config", "", "path to .danger.yaml or .danger.yml")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, path, err := config.Load(*configPath)
	if err != nil {
		return err
	}

	var repo git.Repository
	var gh *github.Client
	var pr *github.PullRequestContext

	if mode == "ci" && github.InActions() {
		gh = github.NewClientFromEnv()
		pr, err = github.ContextFromActions()
		if err != nil {
			return err
		}
		repo, err = gh.Repository(ctx, *pr)
		if err != nil {
			return err
		}
		localRepo, err := git.Inspect(ctx, ".")
		if err == nil {
			repo.Files = localRepo.Files
			repo.UntrackedFiles = localRepo.UntrackedFiles
		}
	} else {
		repo, err = git.Inspect(ctx, ".")
		if err != nil {
			return err
		}
	}

	report := danger.Evaluate(cfg, repo)
	fmt.Fprintf(stdout, "Loaded config: %s\n", path)
	printReport(stdout, report)

	if mode == "ci" && gh != nil && pr != nil {
		if err := gh.PostReportComment(ctx, *pr, report); err != nil {
			fmt.Fprintf(stdout, "Skipping GitHub comment: %s\n", err)
		}
	}

	if report.HasFailures() {
		return fmt.Errorf("danger checks failed")
	}
	return nil
}

func printReport(w io.Writer, report danger.Report) {
	if len(report.Messages) == 0 {
		fmt.Fprintln(w, "No issues found.")
		return
	}

	for _, message := range report.Messages {
		fmt.Fprintf(w, "%s: %s\n", message.Level, message.Text)
	}
}

func init() {
	flag.CommandLine.SetOutput(os.Stderr)
}
