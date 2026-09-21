package config

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

var defaultPaths = []string{".danger.yaml", ".danger.yml"}

type Config struct {
	Level   string   `yaml:"level"`
	Rules   Rules    `yaml:"rules"`
	Plugins []Plugin `yaml:"plugins"`
}

type Rules struct {
	MaxChangedFiles            IntRule        `yaml:"max_changed_files"`
	MaxChangedLines            IntRule        `yaml:"max_changed_lines"`
	RequirePRTitlePattern      StringRule     `yaml:"require_pr_title_pattern"`
	RequireLinkedIssuePattern  StringRule     `yaml:"require_linked_issue_pattern"`
	RequireConventionalCommits BoolRule       `yaml:"require_conventional_commits"`
	RequireSquashedCommits     SquashRule     `yaml:"require_squashed_commits"`
	RequiredLabels             StringListRule `yaml:"required_labels"`
	RequiredFiles              StringListRule `yaml:"required_files"`
	RequiredChangedFiles       StringListRule `yaml:"required_changed_files"`
	ForbiddenFiles             StringListRule `yaml:"forbidden_files"`
	WarnFiles                  StringListRule `yaml:"warn_files"`
	WarnDependencyChanges      BoolRule       `yaml:"warn_dependency_changes"`
}

type IntRule struct {
	Value int
	Level string
}

type StringRule struct {
	Value string
	Level string
}

type BoolRule struct {
	Enabled bool
	Level   string
}

type StringListRule struct {
	Values []string
	Level  string
}

type SquashRule struct {
	Enabled bool
	Level   string
}

type Plugin struct {
	Name    string         `yaml:"name"`
	Command Command        `yaml:"command"`
	Level   string         `yaml:"level"`
	Timeout string         `yaml:"timeout"`
	Config  map[string]any `yaml:"config"`
}

type Command []string

func (r *IntRule) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		return value.Decode(&r.Value)
	case yaml.MappingNode:
		var raw struct {
			Value int    `yaml:"value"`
			Level string `yaml:"level"`
		}
		if err := value.Decode(&raw); err != nil {
			return err
		}
		r.Value = raw.Value
		r.Level = raw.Level
		return nil
	default:
		return fmt.Errorf("expected integer or rule object")
	}
}

func (r *StringRule) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		return value.Decode(&r.Value)
	case yaml.MappingNode:
		var raw struct {
			Value string `yaml:"value"`
			Level string `yaml:"level"`
		}
		if err := value.Decode(&raw); err != nil {
			return err
		}
		r.Value = raw.Value
		r.Level = raw.Level
		return nil
	default:
		return fmt.Errorf("expected string or rule object")
	}
}

func (r *BoolRule) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		return value.Decode(&r.Enabled)
	case yaml.MappingNode:
		var raw struct {
			Enabled bool   `yaml:"enabled"`
			Level   string `yaml:"level"`
		}
		if err := value.Decode(&raw); err != nil {
			return err
		}
		r.Enabled = raw.Enabled
		r.Level = raw.Level
		return nil
	default:
		return fmt.Errorf("expected boolean or rule object")
	}
}

func (r *StringListRule) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.SequenceNode:
		return value.Decode(&r.Values)
	case yaml.MappingNode:
		var raw struct {
			Values []string `yaml:"values"`
			Level  string   `yaml:"level"`
		}
		if err := value.Decode(&raw); err != nil {
			return err
		}
		r.Values = raw.Values
		r.Level = raw.Level
		return nil
	default:
		return fmt.Errorf("expected string list or rule object")
	}
}

func (c *Command) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		var command string
		if err := value.Decode(&command); err != nil {
			return err
		}
		*c = []string{command}
		return nil
	case yaml.SequenceNode:
		var command []string
		if err := value.Decode(&command); err != nil {
			return err
		}
		*c = command
		return nil
	default:
		return fmt.Errorf("expected string or string list")
	}
}

func (r *SquashRule) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		var level string
		if err := value.Decode(&level); err != nil {
			return err
		}
		r.Enabled = level != ""
		r.Level = level
		return nil
	case yaml.MappingNode:
		var raw struct {
			Enabled bool   `yaml:"enabled"`
			Level   string `yaml:"level"`
		}
		if err := value.Decode(&raw); err != nil {
			return err
		}
		r.Enabled = raw.Enabled
		r.Level = raw.Level
		return nil
	default:
		return fmt.Errorf("expected level string or rule object")
	}
}

func Load(path string) (Config, string, error) {
	if path != "" {
		cfg, err := loadPath(path)
		return cfg, path, err
	}

	path, err := FindPath()
	if err != nil {
		return Config{}, "", err
	}
	cfg, err := loadPath(path)
	return cfg, path, err
}

func FindPath() (string, error) {
	for _, candidate := range defaultPaths {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		} else if !errors.Is(err, os.ErrNotExist) {
			return "", err
		}
	}

	return "", fmt.Errorf("no config found; create .danger.yaml or .danger.yml")
}

func loadPath(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c Config) Validate() error {
	if err := validateLevel("level", c.Level, true); err != nil {
		return err
	}

	checks := map[string]string{
		"rules.max_changed_files.level":            c.Rules.MaxChangedFiles.Level,
		"rules.max_changed_lines.level":            c.Rules.MaxChangedLines.Level,
		"rules.require_pr_title_pattern.level":     c.Rules.RequirePRTitlePattern.Level,
		"rules.require_linked_issue_pattern.level": c.Rules.RequireLinkedIssuePattern.Level,
		"rules.require_conventional_commits.level": c.Rules.RequireConventionalCommits.Level,
		"rules.require_squashed_commits.level":     c.Rules.RequireSquashedCommits.Level,
		"rules.required_labels.level":              c.Rules.RequiredLabels.Level,
		"rules.required_files.level":               c.Rules.RequiredFiles.Level,
		"rules.required_changed_files.level":       c.Rules.RequiredChangedFiles.Level,
		"rules.forbidden_files.level":              c.Rules.ForbiddenFiles.Level,
		"rules.warn_files.level":                   c.Rules.WarnFiles.Level,
		"rules.warn_dependency_changes.level":      c.Rules.WarnDependencyChanges.Level,
	}
	for name, level := range checks {
		if err := validateLevel(name, level, true); err != nil {
			return err
		}
	}
	for i, plugin := range c.Plugins {
		prefix := fmt.Sprintf("plugins[%d]", i)
		if plugin.Name == "" {
			return fmt.Errorf("%s.name is required", prefix)
		}
		if len(plugin.Command) == 0 {
			return fmt.Errorf("%s.command is required", prefix)
		}
		if err := validateLevel(prefix+".level", plugin.Level, true); err != nil {
			return err
		}
	}
	return nil
}

func validateLevel(name, level string, allowEmpty bool) error {
	if level == "" && allowEmpty {
		return nil
	}
	if level == "warn" || level == "fail" {
		return nil
	}
	return fmt.Errorf("%s must be \"warn\" or \"fail\", got %q", name, level)
}
