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
	if cfg.Rules.MaxChangedFiles != 12 {
		t.Fatalf("max changed files = %d, want 12", cfg.Rules.MaxChangedFiles)
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
	if cfg.Rules.MaxChangedFiles != 7 {
		t.Fatalf("max changed files = %d, want 7", cfg.Rules.MaxChangedFiles)
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
	if cfg.Rules.MaxChangedFiles != 3 {
		t.Fatalf("max changed files = %d, want 3", cfg.Rules.MaxChangedFiles)
	}
}

func TestLoadCommitPolicyRules(t *testing.T) {
	t.Chdir(t.TempDir())
	writeFile(t, ".danger.yaml", "rules:\n  require_conventional_commits: true\n  require_squashed_commits: warn\n")

	cfg, _, err := Load("")
	if err != nil {
		t.Fatal(err)
	}

	if !cfg.Rules.RequireConventionalCommits {
		t.Fatal("require_conventional_commits = false, want true")
	}
	if cfg.Rules.RequireSquashedCommits != "warn" {
		t.Fatalf("require_squashed_commits = %q, want warn", cfg.Rules.RequireSquashedCommits)
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
