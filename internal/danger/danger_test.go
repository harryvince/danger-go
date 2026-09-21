package danger

import (
	"testing"

	"github.com/harryvince/danger-go/internal/config"
	"github.com/harryvince/danger-go/internal/git"
)

func TestEvaluateReportsFailures(t *testing.T) {
	report := Evaluate(config.Config{
		Rules: config.Rules{
			MaxChangedFiles:           1,
			MaxChangedLines:           5,
			RequirePRTitlePattern:     "^JIRA-\\d+: .+",
			RequireLinkedIssuePattern: "JIRA-\\d+",
			RequiredLabels:            []string{"ready"},
			RequiredFiles:             []string{"README.md"},
			RequiredChangedFiles:      []string{"docs/**"},
			ForbiddenFiles:            []string{"*.tmp"},
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
	if len(report.Messages) != 8 {
		t.Fatalf("message count = %d, want 8", len(report.Messages))
	}
}

func TestEvaluatePassesLabelAndLinkedIssueRules(t *testing.T) {
	report := Evaluate(config.Config{
		Rules: config.Rules{
			RequireLinkedIssuePattern: "JIRA-\\d+",
			RequiredLabels:            []string{"ready"},
		},
	}, git.Repository{
		PullRequestTitle:  "Update workflow",
		PullRequestBody:   "Closes JIRA-123",
		PullRequestLabels: []string{"Ready"},
	})

	if report.HasFailures() {
		t.Fatalf("expected no failures, got %#v", report.Messages)
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

func TestEvaluateRequiresConventionalCommits(t *testing.T) {
	report := Evaluate(config.Config{
		Rules: config.Rules{
			RequireConventionalCommits: true,
		},
	}, git.Repository{
		Commits: []git.Commit{
			{Subject: "feat: add checkout validation"},
			{Subject: "JIRA-123: add checkout validation"},
		},
	})

	if !report.HasFailures() {
		t.Fatal("expected non-conventional commit to fail")
	}
	if got, want := len(report.Messages), 1; got != want {
		t.Fatalf("message count = %d, want %d", got, want)
	}
}

func TestEvaluateWarnsWhenSquashRequiredAsWarning(t *testing.T) {
	report := Evaluate(config.Config{
		Rules: config.Rules{
			RequireSquashedCommits: "warn",
		},
	}, git.Repository{
		Commits: []git.Commit{
			{Subject: "feat: add checkout validation"},
			{Subject: "test: cover checkout validation"},
		},
	})

	if report.HasFailures() {
		t.Fatal("squash warning should not fail the report")
	}
	if got, want := len(report.Messages), 1; got != want {
		t.Fatalf("message count = %d, want %d", got, want)
	}
	if report.Messages[0].Level != LevelWarn {
		t.Fatalf("message level = %q, want warn", report.Messages[0].Level)
	}
}

func TestEvaluateFailsWhenSquashRequiredAsFailure(t *testing.T) {
	report := Evaluate(config.Config{
		Rules: config.Rules{
			RequireSquashedCommits: "fail",
		},
	}, git.Repository{
		Commits: []git.Commit{
			{Subject: "feat: add checkout validation"},
			{Subject: "test: cover checkout validation"},
		},
	})

	if !report.HasFailures() {
		t.Fatal("expected multiple commits to fail")
	}
}

func TestEvaluateRejectsInvalidSquashMode(t *testing.T) {
	report := Evaluate(config.Config{
		Rules: config.Rules{
			RequireSquashedCommits: "deny",
		},
	}, git.Repository{})

	if !report.HasFailures() {
		t.Fatal("expected invalid squash mode to fail")
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
