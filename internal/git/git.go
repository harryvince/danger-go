package git

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"
)

type Repository struct {
	Files            []string
	ModifiedFiles    []string
	UntrackedFiles   []string
	PullRequestTitle string
}

func Inspect(ctx context.Context, dir string) (Repository, error) {
	files, err := gitLines(ctx, dir, "ls-files")
	if err != nil {
		return Repository{}, err
	}

	modified, err := changedSinceHead(ctx, dir)
	if err != nil {
		return Repository{}, err
	}

	untracked, err := gitLines(ctx, dir, "ls-files", "--others", "--exclude-standard")
	if err != nil {
		return Repository{}, err
	}

	return Repository{
		Files:            files,
		ModifiedFiles:    modified,
		UntrackedFiles:   untracked,
		PullRequestTitle: os.Getenv("DANGER_PR_TITLE"),
	}, nil
}

func changedSinceHead(ctx context.Context, dir string) ([]string, error) {
	if _, err := gitLines(ctx, dir, "rev-parse", "--verify", "HEAD"); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return gitLines(ctx, dir, "ls-files")
		}
		return nil, err
	}

	return gitLines(ctx, dir, "diff", "--name-only", "HEAD")
}

func (r Repository) ChangedFiles() []string {
	seen := map[string]bool{}
	var files []string
	for _, file := range append(r.ModifiedFiles, r.UntrackedFiles...) {
		if !seen[file] {
			files = append(files, file)
			seen[file] = true
		}
	}
	return files
}

func (r Repository) HasFile(path string) bool {
	for _, file := range r.Files {
		if file == path {
			return true
		}
	}
	for _, file := range r.UntrackedFiles {
		if file == path {
			return true
		}
	}
	return false
}

func gitLines(ctx context.Context, dir string, args ...string) ([]string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	text := strings.TrimSpace(string(output))
	if text == "" {
		return nil, nil
	}
	return strings.Split(text, "\n"), nil
}
