package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadPrefersDangerYAML(t *testing.T) {
	t.Chdir(t.TempDir())
	writeFile(t, ".danger.yaml", "rules:\n  max_changed_files: 12\n")
	writeFile(t, ".danger.yml", "rules:\n  max_changed_files: 99\n")

	cfg, path, err := Load("")
	if err != nil {
		t.Fatal(err)
	}

	if path != ".danger.yaml" {
		t.Fatalf("path = %q, want .danger.yaml", path)
	}
	if cfg.Rules.MaxChangedFiles.Value != 12 {
		t.Fatalf("max changed files = %d, want 12", cfg.Rules.MaxChangedFiles.Value)
	}
}

func TestLoadFallsBackToDangerYML(t *testing.T) {
	t.Chdir(t.TempDir())
	writeFile(t, ".danger.yml", "rules:\n  max_changed_files: 7\n")

	cfg, path, err := Load("")
	if err != nil {
		t.Fatal(err)
	}

	if path != ".danger.yml" {
		t.Fatalf("path = %q, want .danger.yml", path)
	}
	if cfg.Rules.MaxChangedFiles.Value != 7 {
		t.Fatalf("max changed files = %d, want 7", cfg.Rules.MaxChangedFiles.Value)
	}
}

func TestLoadExplicitPath(t *testing.T) {
	t.Chdir(t.TempDir())
	path := filepath.Join("configs", "danger.yaml")
	writeFile(t, path, "rules:\n  max_changed_files: 3\n")

	cfg, loadedPath, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	if loadedPath != path {
		t.Fatalf("path = %q, want %q", loadedPath, path)
	}
	if cfg.Rules.MaxChangedFiles.Value != 3 {
		t.Fatalf("max changed files = %d, want 3", cfg.Rules.MaxChangedFiles.Value)
	}
}

func TestLoadCommitPolicyRules(t *testing.T) {
	t.Chdir(t.TempDir())
	writeFile(t, ".danger.yaml", "rules:\n  require_conventional_commits: true\n  require_signed_off_commits: true\n  require_squashed_commits: warn\n")

	cfg, _, err := Load("")
	if err != nil {
		t.Fatal(err)
	}

	if !cfg.Rules.RequireConventionalCommits.Enabled {
		t.Fatal("require_conventional_commits = false, want true")
	}
	if !cfg.Rules.RequireSignedOffCommits.Enabled {
		t.Fatal("require_signed_off_commits = false, want true")
	}
	if !cfg.Rules.RequireSquashedCommits.Enabled {
		t.Fatal("require_squashed_commits disabled, want enabled")
	}
	if cfg.Rules.RequireSquashedCommits.Level != "warn" {
		t.Fatalf("require_squashed_commits level = %q, want warn", cfg.Rules.RequireSquashedCommits.Level)
	}
}

func TestLoadObjectRulesAndTopLevel(t *testing.T) {
	t.Chdir(t.TempDir())
	writeFile(t, ".danger.yaml", `level: warn
rules:
  max_changed_files:
    value: 12
    level: fail
  require_pr_title_pattern:
    value: "^ISSUE-[0-9]+: .+"
    level: warn
  required_labels:
    values:
      - ready
    level: fail
  warn_dependency_changes:
    enabled: true
    level: fail
`)

	cfg, _, err := Load("")
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Level != "warn" {
		t.Fatalf("level = %q, want warn", cfg.Level)
	}
	if cfg.Rules.MaxChangedFiles.Value != 12 || cfg.Rules.MaxChangedFiles.Level != "fail" {
		t.Fatalf("max_changed_files = %#v", cfg.Rules.MaxChangedFiles)
	}
	if cfg.Rules.RequirePRTitlePattern.Value != "^ISSUE-[0-9]+: .+" || cfg.Rules.RequirePRTitlePattern.Level != "warn" {
		t.Fatalf("require_pr_title_pattern = %#v", cfg.Rules.RequirePRTitlePattern)
	}
	if got := cfg.Rules.RequiredLabels.Values; len(got) != 1 || got[0] != "ready" || cfg.Rules.RequiredLabels.Level != "fail" {
		t.Fatalf("required_labels = %#v", cfg.Rules.RequiredLabels)
	}
	if !cfg.Rules.WarnDependencyChanges.Enabled || cfg.Rules.WarnDependencyChanges.Level != "fail" {
		t.Fatalf("warn_dependency_changes = %#v", cfg.Rules.WarnDependencyChanges)
	}
}

func TestLabelsDefaultEnabled(t *testing.T) {
	t.Chdir(t.TempDir())
	writeFile(t, ".danger.yaml", "rules:\n  max_changed_files: 3\n")

	cfg, _, err := Load("")
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Labels.IsEnabled() {
		t.Fatal("labels disabled, want enabled by default")
	}
}

func TestLabelsCanBeDisabled(t *testing.T) {
	t.Chdir(t.TempDir())
	writeFile(t, ".danger.yaml", "labels: false\nrules:\n  max_changed_files: 3\n")

	cfg, _, err := Load("")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Labels.IsEnabled() {
		t.Fatal("labels enabled, want disabled")
	}
}

func TestLabelsObjectCanBeDisabled(t *testing.T) {
	t.Chdir(t.TempDir())
	writeFile(t, ".danger.yaml", "labels:\n  enabled: false\nrules:\n  max_changed_files: 3\n")

	cfg, _, err := Load("")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Labels.IsEnabled() {
		t.Fatal("labels enabled, want disabled")
	}
}

func TestLoadPlugins(t *testing.T) {
	t.Chdir(t.TempDir())
	writeFile(t, ".danger.yaml", `plugins:
  - name: custom
    command:
      - go
      - run
      - ./tools/policy
    level: warn
    timeout: 10s
    config:
      owner: platform
`)

	cfg, _, err := Load("")
	if err != nil {
		t.Fatal(err)
	}

	if got, want := len(cfg.Plugins), 1; got != want {
		t.Fatalf("plugin count = %d, want %d", got, want)
	}
	plugin := cfg.Plugins[0]
	if plugin.Name != "custom" || plugin.Level != "warn" || plugin.Timeout != "10s" {
		t.Fatalf("plugin metadata = %#v", plugin)
	}
	if got := []string(plugin.Command); len(got) != 3 || got[0] != "go" || got[2] != "./tools/policy" {
		t.Fatalf("plugin command = %#v", got)
	}
	if plugin.Config["owner"] != "platform" {
		t.Fatalf("plugin config = %#v", plugin.Config)
	}
}

func TestLoadRejectsInvalidLevel(t *testing.T) {
	t.Chdir(t.TempDir())
	writeFile(t, ".danger.yaml", "level: info\nrules:\n  max_changed_files: 3\n")

	if _, _, err := Load(""); err == nil {
		t.Fatal("expected invalid level error")
	}
}

func TestLoadRejectsInvalidRuleLevel(t *testing.T) {
	t.Chdir(t.TempDir())
	writeFile(t, ".danger.yaml", "rules:\n  max_changed_files:\n    value: 3\n    level: info\n")

	if _, _, err := Load(""); err == nil {
		t.Fatal("expected invalid rule level error")
	}
}

func writeFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}
