package git

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type FileChange struct {
	Path      string
	Additions int
	Deletions int
}

type Repository struct {
	Files             []string
	ModifiedFiles     []string
	UntrackedFiles    []string
	FileChanges       []FileChange
	PullRequestTitle  string
	PullRequestBody   string
	PullRequestBranch string
	PullRequestLabels []string
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

	fileChanges, err := inspectFileChanges(ctx, dir, modified, untracked)
	if err != nil {
		return Repository{}, err
	}

	return Repository{
		Files:             files,
		ModifiedFiles:     modified,
		UntrackedFiles:    untracked,
		FileChanges:       fileChanges,
		PullRequestTitle:  os.Getenv("DANGER_PR_TITLE"),
		PullRequestBody:   os.Getenv("DANGER_PR_BODY"),
		PullRequestBranch: os.Getenv("DANGER_PR_BRANCH"),
		PullRequestLabels: splitEnvList(os.Getenv("DANGER_PR_LABELS")),
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
	for _, change := range r.FileChanges {
		if !seen[change.Path] {
			files = append(files, change.Path)
			seen[change.Path] = true
		}
	}
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

func (r Repository) ChangedLines() int {
	var lines int
	for _, change := range r.FileChanges {
		lines += change.Additions + change.Deletions
	}
	return lines
}

func inspectFileChanges(ctx context.Context, dir string, modified, untracked []string) ([]FileChange, error) {
	lines, err := gitLines(ctx, dir, "diff", "--numstat", "HEAD")
	if err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) {
			return nil, err
		}
		lines, err = gitLines(ctx, dir, "diff", "--numstat", "--cached")
		if err != nil {
			return nil, err
		}
	}

	changes := make([]FileChange, 0, len(lines)+len(untracked))
	seen := map[string]bool{}
	for _, line := range lines {
		parts := strings.Split(line, "\t")
		if len(parts) < 3 {
			continue
		}
		path := parts[2]
		additions := parseNumstat(parts[0])
		deletions := parseNumstat(parts[1])
		changes = append(changes, FileChange{Path: path, Additions: additions, Deletions: deletions})
		seen[path] = true
	}

	for _, file := range append(modified, untracked...) {
		if seen[file] {
			continue
		}
		changes = append(changes, FileChange{Path: file})
	}
	return changes, nil
}

func parseNumstat(value string) int {
	n, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return n
}

func splitEnvList(value string) []string {
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	var values []string
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			values = append(values, part)
		}
	}
	return values
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
