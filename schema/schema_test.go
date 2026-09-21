package schema

import "testing"

func TestValidateAcceptsValidConfig(t *testing.T) {
	data := []byte(`
$schema: https://harryvince.github.io/danger-go/schema/danger-go.schema.json
rules:
  max_changed_files: 10
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

func TestValidateRejectsWrongType(t *testing.T) {
	data := []byte(`
rules:
  max_changed_files: many
`)

	if err := Validate(data); err == nil {
		t.Fatal("expected validation error")
	}
}
