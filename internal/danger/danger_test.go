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
			RequirePRTitlePattern: "^JIRA-\\d+: .+",
			RequiredFiles:         []string{"README.md"},
			ForbiddenFiles:        []string{"*.tmp"},
		},
	}, git.Repository{
		Files:            []string{"go.mod"},
		ModifiedFiles:    []string{"main.go", "scratch.tmp"},
		PullRequestTitle: "missing ticket",
	})

	if !report.HasFailures() {
		t.Fatal("expected failures")
	}
	if len(report.Messages) != 4 {
		t.Fatalf("message count = %d, want 4", len(report.Messages))
	}
}
