package danger

import (
	"testing"

	"github.com/harryvince/danger-go/internal/config"
	"github.com/harryvince/danger-go/internal/git"
)

func TestEvaluateReportsFailures(t *testing.T) {
	report := Evaluate(config.Config{
		Rules: config.Rules{
			MaxChangedFiles:           config.IntRule{Value: 1},
			MaxChangedLines:           config.IntRule{Value: 5},
			RequirePRTitlePattern:     config.StringRule{Value: "^ISSUE-\\d+: .+"},
			RequireLinkedIssuePattern: config.StringRule{Value: "ISSUE-\\d+"},
			RequiredLabels:            config.StringListRule{Values: []string{"ready"}},
			RequiredFiles:             config.StringListRule{Values: []string{"README.md"}},
			RequiredChangedFiles:      config.StringListRule{Values: []string{"docs/**"}},
			ForbiddenFiles:            config.StringListRule{Values: []string{"*.tmp"}},
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
			RequireLinkedIssuePattern: config.StringRule{Value: "ISSUE-\\d+"},
			RequiredLabels:            config.StringListRule{Values: []string{"ready"}},
		},
	}, git.Repository{
		PullRequestTitle:  "Update workflow",
		PullRequestBody:   "Closes ISSUE-123",
		PullRequestLabels: []string{"Ready"},
	})

	if report.HasFailures() {
		t.Fatalf("expected no failures, got %#v", report.Messages)
	}
}

func TestEvaluateReportsWarnings(t *testing.T) {
	report := Evaluate(config.Config{
		Rules: config.Rules{
			WarnFiles:             config.StringListRule{Values: []string{"generated/**"}},
			WarnDependencyChanges: config.BoolRule{Enabled: true},
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
			RequireConventionalCommits: config.BoolRule{Enabled: true},
		},
	}, git.Repository{
		Commits: []git.Commit{
			{Subject: "feat: add checkout validation"},
			{Subject: "ISSUE-123: add checkout validation"},
		},
	})

	if !report.HasFailures() {
		t.Fatal("expected non-conventional commit to fail")
	}
	if got, want := len(report.Messages), 1; got != want {
		t.Fatalf("message count = %d, want %d", got, want)
	}
}

func TestEvaluateRequiresSignedOffCommits(t *testing.T) {
	report := Evaluate(config.Config{
		Rules: config.Rules{
			RequireSignedOffCommits: config.BoolRule{Enabled: true},
		},
	}, git.Repository{
		Commits: []git.Commit{
			{
				Subject: "feat: add checkout validation",
				Message: "feat: add checkout validation\n\nSigned-off-by: Mona Lisa <mona@example.com>",
			},
			{
				Subject: "test: cover checkout validation",
				Message: "test: cover checkout validation",
			},
		},
	})

	if !report.HasFailures() {
		t.Fatal("expected unsigned commit to fail")
	}
	if got, want := len(report.Messages), 1; got != want {
		t.Fatalf("message count = %d, want %d", got, want)
	}
}

func TestEvaluateWarnsWhenSquashRequiredAsWarning(t *testing.T) {
	report := Evaluate(config.Config{
		Rules: config.Rules{
			RequireSquashedCommits: config.SquashRule{Enabled: true, Level: "warn"},
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
			RequireSquashedCommits: config.SquashRule{Enabled: true, Level: "fail"},
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

func TestEvaluateUsesTopLevelRuleLevel(t *testing.T) {
	report := Evaluate(config.Config{
		Level: "warn",
		Rules: config.Rules{
			MaxChangedFiles: config.IntRule{Value: 1},
		},
	}, git.Repository{
		ModifiedFiles: []string{"one.go", "two.go"},
	})

	if report.HasFailures() {
		t.Fatal("top-level warning should not fail the report")
	}
	if got, want := len(report.Messages), 1; got != want {
		t.Fatalf("message count = %d, want %d", got, want)
	}
	if report.Messages[0].Level != LevelWarn {
		t.Fatalf("message level = %q, want warn", report.Messages[0].Level)
	}
}

func TestEvaluateUsesRuleLevelOverride(t *testing.T) {
	report := Evaluate(config.Config{
		Level: "warn",
		Rules: config.Rules{
			MaxChangedFiles: config.IntRule{Value: 1, Level: "fail"},
		},
	}, git.Repository{
		ModifiedFiles: []string{"one.go", "two.go"},
	})

	if !report.HasFailures() {
		t.Fatal("rule-level failure should fail the report")
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
