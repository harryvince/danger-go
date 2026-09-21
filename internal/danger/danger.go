package danger

import (
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/harryvince/danger-go/internal/config"
	"github.com/harryvince/danger-go/internal/git"
)

var dependencyManifestPatterns = []string{
	"go.mod",
	"go.sum",
	"package.json",
	"package-lock.json",
	"pnpm-lock.yaml",
	"yarn.lock",
	"Gemfile",
	"Gemfile.lock",
	"Cargo.toml",
	"Cargo.lock",
	"pyproject.toml",
	"poetry.lock",
	"requirements.txt",
}

var conventionalCommitPattern = regexp.MustCompile(`^(build|chore|ci|docs|feat|fix|perf|refactor|revert|style|test)(\([^)]+\))?!?: .+`)

type Level string

const (
	LevelFail Level = "fail"
	LevelWarn Level = "warn"
)

type Message struct {
	Level Level
	Text  string
}

type Report struct {
	Messages []Message
}

func (r Report) HasFailures() bool {
	for _, message := range r.Messages {
		if message.Level == LevelFail {
			return true
		}
	}
	return false
}

func Evaluate(cfg config.Config, repo git.Repository) Report {
	var report Report

	changedFiles := repo.ChangedFiles()
	if cfg.Rules.MaxChangedFiles > 0 && len(changedFiles) > cfg.Rules.MaxChangedFiles {
		report.fail(fmt.Sprintf("PR changes %d files, above the configured limit of %d", len(changedFiles), cfg.Rules.MaxChangedFiles))
	}

	changedLines := repo.ChangedLines()
	if cfg.Rules.MaxChangedLines > 0 && changedLines > cfg.Rules.MaxChangedLines {
		report.fail(fmt.Sprintf("PR changes %d lines, above the configured limit of %d", changedLines, cfg.Rules.MaxChangedLines))
	}

	if cfg.Rules.RequirePRTitlePattern != "" {
		matched, err := regexp.MatchString(cfg.Rules.RequirePRTitlePattern, repo.PullRequestTitle)
		if err != nil {
			report.fail(fmt.Sprintf("invalid require_pr_title_pattern: %s", err))
		} else if !matched {
			report.fail(fmt.Sprintf("PR title %q does not match %q", repo.PullRequestTitle, cfg.Rules.RequirePRTitlePattern))
		}
	}

	if cfg.Rules.RequireLinkedIssuePattern != "" {
		text := strings.Join([]string{repo.PullRequestTitle, repo.PullRequestBody, repo.PullRequestBranch}, "\n")
		matched, err := regexp.MatchString(cfg.Rules.RequireLinkedIssuePattern, text)
		if err != nil {
			report.fail(fmt.Sprintf("invalid require_linked_issue_pattern: %s", err))
		} else if !matched {
			report.fail(fmt.Sprintf("PR title, body, or branch does not match linked issue pattern %q", cfg.Rules.RequireLinkedIssuePattern))
		}
	}

	if cfg.Rules.RequireConventionalCommits {
		for _, commit := range repo.Commits {
			if !conventionalCommitPattern.MatchString(commit.Subject) {
				report.fail(fmt.Sprintf("commit subject %q is not conventional", commit.Subject))
			}
		}
	}

	switch cfg.Rules.RequireSquashedCommits {
	case "fail":
		if len(repo.Commits) > 1 {
			report.fail(fmt.Sprintf("PR has %d commits; squash to a single commit before merging", len(repo.Commits)))
		}
	case "warn":
		if len(repo.Commits) > 1 {
			report.warn(fmt.Sprintf("PR has %d commits; consider squashing before merging", len(repo.Commits)))
		}
	case "":
	default:
		report.fail(fmt.Sprintf("invalid require_squashed_commits value %q; use \"warn\" or \"fail\"", cfg.Rules.RequireSquashedCommits))
	}

	for _, label := range cfg.Rules.RequiredLabels {
		if !hasLabel(repo.PullRequestLabels, label) {
			report.fail(fmt.Sprintf("required label is missing: %s", label))
		}
	}

	for _, required := range cfg.Rules.RequiredFiles {
		if !repo.HasFile(required) {
			report.fail(fmt.Sprintf("required file is missing: %s", required))
		}
	}

	for _, required := range cfg.Rules.RequiredChangedFiles {
		if !hasChangedFile(changedFiles, required) {
			report.fail(fmt.Sprintf("required changed file pattern is missing: %s", required))
		}
	}

	for _, forbidden := range cfg.Rules.ForbiddenFiles {
		for _, changed := range changedFiles {
			if matchedPath(forbidden, changed) {
				report.fail(fmt.Sprintf("forbidden file changed: %s", changed))
			}
		}
	}

	for _, warning := range cfg.Rules.WarnFiles {
		for _, changed := range changedFiles {
			if matchedPath(warning, changed) {
				report.warn(fmt.Sprintf("watched file changed: %s", changed))
			}
		}
	}

	if cfg.Rules.WarnDependencyChanges {
		for _, changed := range changedFiles {
			if matchesAny(dependencyManifestPatterns, changed) {
				report.warn(fmt.Sprintf("dependency manifest changed: %s", changed))
			}
		}
	}

	return report
}

func (r *Report) fail(text string) {
	r.Messages = append(r.Messages, Message{Level: LevelFail, Text: text})
}

func (r *Report) warn(text string) {
	r.Messages = append(r.Messages, Message{Level: LevelWarn, Text: text})
}

func hasChangedFile(changedFiles []string, pattern string) bool {
	for _, changed := range changedFiles {
		if matchedPath(pattern, changed) {
			return true
		}
	}
	return false
}

func matchesAny(patterns []string, changed string) bool {
	for _, pattern := range patterns {
		if matchedPath(pattern, changed) {
			return true
		}
	}
	return false
}

func hasLabel(labels []string, required string) bool {
	for _, label := range labels {
		if strings.EqualFold(label, required) {
			return true
		}
	}
	return false
}

func matchedPath(pattern, path string) bool {
	normalizedPattern := normalizePath(pattern)
	normalizedPath := normalizePath(path)

	if normalizedPattern == normalizedPath {
		return true
	}

	if strings.HasSuffix(normalizedPattern, "/**") {
		prefix := strings.TrimSuffix(normalizedPattern, "/**")
		return normalizedPath == prefix || strings.HasPrefix(normalizedPath, prefix+"/")
	}

	matched, err := pathpkgMatch(normalizedPattern, normalizedPath)
	return err == nil && matched
}

func normalizePath(value string) string {
	return strings.TrimPrefix(strings.ReplaceAll(value, "\\", "/"), "./")
}

func pathpkgMatch(pattern, name string) (bool, error) {
	return path.Match(pattern, name)
}
