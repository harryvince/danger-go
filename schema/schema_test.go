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

func TestValidateRejectsUnknownRule(t *testing.T) {
	data := []byte(`
rules:
  made_up_rule: true
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
