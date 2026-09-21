package schema

import "testing"

func TestValidateAcceptsValidConfig(t *testing.T) {
	data := []byte(`
$schema: https://harryvince.github.io/danger-go/schema/danger-go.schema.json
rules:
  max_changed_files: 10
  require_conventional_commits: true
  require_squashed_commits: warn
  required_labels:
    - ready
  warn_dependency_changes: true
`)

	if err := Validate(data); err != nil {
		t.Fatal(err)
	}
}

func TestValidateAcceptsObjectRules(t *testing.T) {
	data := []byte(`
level: warn
rules:
  max_changed_files:
    value: 10
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

	if err := Validate(data); err != nil {
		t.Fatal(err)
	}
}

func TestValidateRejectsUnknownRule(t *testing.T) {
	data := []byte(`
rules:
  made_up_rule: true
`)

	if err := Validate(data); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestValidateRejectsInvalidTopLevelLevel(t *testing.T) {
	data := []byte(`
level: info
rules:
  max_changed_files: 10
`)

	if err := Validate(data); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestValidateRejectsInvalidRuleLevel(t *testing.T) {
	data := []byte(`
rules:
  max_changed_files:
    value: 10
    level: info
`)

	if err := Validate(data); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestValidateRejectsInvalidSquashMode(t *testing.T) {
	data := []byte(`
rules:
  require_squashed_commits: deny
`)

	if err := Validate(data); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestValidateRejectsWrongType(t *testing.T) {
	data := []byte(`
rules:
  max_changed_files: many
`)

	if err := Validate(data); err == nil {
		t.Fatal("expected validation error")
	}
}
