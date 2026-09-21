package schema

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"gopkg.in/yaml.v3"
)

//go:embed danger-go.schema.json
var schemaData []byte

func ValidateFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return Validate(data)
}

func Validate(data []byte) error {
	var yamlValue any
	if err := yaml.Unmarshal(data, &yamlValue); err != nil {
		return err
	}

	var schemaValue any
	if err := json.Unmarshal(schemaData, &schemaValue); err != nil {
		return err
	}

	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource("danger-go.schema.json", schemaValue); err != nil {
		return err
	}

	schema, err := compiler.Compile("danger-go.schema.json")
	if err != nil {
		return err
	}

	if err := schema.Validate(normalizeYAML(yamlValue)); err != nil {
		return fmt.Errorf("config does not match schema: %w", err)
	}
	return nil
}

func normalizeYAML(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		result := make(map[string]any, len(typed))
		for key, value := range typed {
			result[key] = normalizeYAML(value)
		}
		return result
	case []any:
		result := make([]any, len(typed))
		for i, value := range typed {
			result[i] = normalizeYAML(value)
		}
		return result
	default:
		return typed
	}
}
