package danger

import (
	"testing"

	"github.com/harryvince/danger-go/internal/config"
	"github.com/harryvince/danger-go/internal/git"
)

func TestEvaluateReportsFailures(t *testing.T) {
	report := Evaluate(config.Config{
		Rules: config.Rules{
			MaxChangedFiles:       1,
			MaxChangedLines:       5,
			RequirePRTitlePattern: "^JIRA-\\d+: .+",
			RequiredFiles:         []string{"README.md"},
			RequiredChangedFiles:  []string{"docs/**"},
			ForbiddenFiles:        []string{"*.tmp"},
		},
	}, git.Repository{
		Files:            []string{"go.mod"},
		ModifiedFiles:    []string{"main.go", "scratch.tmp"},
		FileChanges:      []git.FileChange{{Path: "main.go", Additions: 4, Deletions: 3}, {Path: "scratch.tmp"}},
		PullRequestTitle: "missing ticket",
	})

	if !report.HasFailures() {
		t.Fatal("expected failures")
	}
	if len(report.Messages) != 6 {
		t.Fatalf("message count = %d, want 6", len(report.Messages))
	}
}

func TestEvaluateReportsWarnings(t *testing.T) {
	report := Evaluate(config.Config{
		Rules: config.Rules{
			WarnFiles:             []string{"generated/**"},
			WarnDependencyChanges: true,
		},
	}, git.Repository{
		FileChanges: []git.FileChange{
			{Path: "generated/client.go"},
			{Path: "go.mod"},
		},
	})

	if report.HasFailures() {
		t.Fatal("warnings should not fail the report")
	}
	if len(report.Messages) != 2 {
		t.Fatalf("message count = %d, want 2", len(report.Messages))
	}
	for _, message := range report.Messages {
		if message.Level != LevelWarn {
			t.Fatalf("message level = %q, want warn", message.Level)
		}
	}
}

func TestMatchedPathSupportsDirectoryGlob(t *testing.T) {
	if !matchedPath("docs/**", "docs/configuration.md") {
		t.Fatal("expected docs/** to match nested docs file")
	}
	if matchedPath("docs/**", "README.md") {
		t.Fatal("expected docs/** not to match README.md")
	}
}
