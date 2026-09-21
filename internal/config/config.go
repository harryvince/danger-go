package config

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

var defaultPaths = []string{".danger.yaml", ".danger.yml"}

type Config struct {
	Rules Rules `yaml:"rules"`
}

type Rules struct {
	MaxChangedFiles            int      `yaml:"max_changed_files"`
	MaxChangedLines            int      `yaml:"max_changed_lines"`
	RequirePRTitlePattern      string   `yaml:"require_pr_title_pattern"`
	RequireLinkedIssuePattern  string   `yaml:"require_linked_issue_pattern"`
	RequireConventionalCommits bool     `yaml:"require_conventional_commits"`
	RequireSquashedCommits     string   `yaml:"require_squashed_commits"`
	RequiredLabels             []string `yaml:"required_labels"`
	RequiredFiles              []string `yaml:"required_files"`
	RequiredChangedFiles       []string `yaml:"required_changed_files"`
	ForbiddenFiles             []string `yaml:"forbidden_files"`
	WarnFiles                  []string `yaml:"warn_files"`
	WarnDependencyChanges      bool     `yaml:"warn_dependency_changes"`
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

	return cfg, nil
}
