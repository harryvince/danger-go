package plugin

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/harryvince/danger-go/internal/config"
	"github.com/harryvince/danger-go/internal/danger"
	"github.com/harryvince/danger-go/internal/git"
)

func TestRunPassesRepoAndConfigToPlugin(t *testing.T) {
	pluginPath := writeScript(t, `#!/bin/sh
python3 -c 'import json,sys; req=json.load(sys.stdin); assert req["protocol_version"] == "v1"; assert req["repo"]["pull_request"]["title"] == "JIRA-123: test"; assert req["config"]["owner"] == "platform"; print(json.dumps({"messages":[{"level":"warn","text":"custom warning"}]}))'
`)

	report := Run(context.Background(), config.Plugin{Name: "custom", Command: []string{pluginPath}, Config: map[string]any{"owner": "platform"}}, git.Repository{PullRequestTitle: "JIRA-123: test"}, "")

	if report.HasFailures() {
		t.Fatalf("expected warning-only report, got %#v", report.Messages)
	}
	if got, want := len(report.Messages), 1; got != want {
		t.Fatalf("message count = %d, want %d", got, want)
	}
	if report.Messages[0].Level != danger.LevelWarn || report.Messages[0].Text != "[custom] custom warning" {
		t.Fatalf("unexpected message: %#v", report.Messages[0])
	}
}

func TestRunReportsNonZeroExit(t *testing.T) {
	pluginPath := writeScript(t, "#!/bin/sh\necho broken >&2\nexit 2\n")

	report := Run(context.Background(), config.Plugin{Name: "bad", Command: []string{pluginPath}}, git.Repository{}, "")

	if !report.HasFailures() {
		t.Fatal("expected plugin failure")
	}
	if !strings.Contains(report.Messages[0].Text, "plugin exited with error") || !strings.Contains(report.Messages[0].Text, "broken") {
		t.Fatalf("unexpected message: %q", report.Messages[0].Text)
	}
}

func TestRunReportsInvalidJSON(t *testing.T) {
	pluginPath := writeScript(t, "#!/bin/sh\necho not-json\n")

	report := Run(context.Background(), config.Plugin{Name: "bad-json", Command: []string{pluginPath}}, git.Repository{}, "")

	if !report.HasFailures() {
		t.Fatal("expected plugin failure")
	}
	if !strings.Contains(report.Messages[0].Text, "invalid JSON") {
		t.Fatalf("unexpected message: %q", report.Messages[0].Text)
	}
}

func TestRunReportsTimeout(t *testing.T) {
	pluginPath := writeScript(t, "#!/bin/sh\nsleep 2\n")

	report := Run(context.Background(), config.Plugin{Name: "slow", Command: []string{pluginPath}, Timeout: "1ms"}, git.Repository{}, "")

	if !report.HasFailures() {
		t.Fatal("expected plugin failure")
	}
	if !strings.Contains(report.Messages[0].Text, "timed out") {
		t.Fatalf("unexpected message: %q", report.Messages[0].Text)
	}
}

func writeScript(t *testing.T, content string) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("shell script tests are not supported on windows")
	}
	path := filepath.Join(t.TempDir(), "plugin.sh")
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}
