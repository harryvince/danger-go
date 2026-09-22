package danger

import (
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/harryvince/danger-go/internal/config"
	"github.com/harryvince/danger-go/internal/git"
)

var dependencyManifestPatterns = []string{
	"go.mod",
	"go.sum",
	"package.json",
	"package-lock.json",
	"pnpm-lock.yaml",
	"yarn.lock",
	"Gemfile",
	"Gemfile.lock",
	"Cargo.toml",
	"Cargo.lock",
	"pyproject.toml",
	"poetry.lock",
	"requirements.txt",
}

var conventionalCommitPattern = regexp.MustCompile(`^(build|chore|ci|docs|feat|fix|perf|refactor|revert|style|test)(\([^)]+\))?!?: .+`)
var signedOffByPattern = regexp.MustCompile(`(?im)^Signed-off-by:\s+\S.+\s+<[^<>@\s]+@[^<>@\s]+>$`)

type Level string

const (
	LevelFail Level = "fail"
	LevelWarn Level = "warn"
)

type Message struct {
	Level Level
	Text  string
}

type Report struct {
	Messages []Message
}

func (r Report) HasFailures() bool {
	for _, message := range r.Messages {
		if message.Level == LevelFail {
			return true
		}
	}
	return false
}

func (r Report) HasWarnings() bool {
	for _, message := range r.Messages {
		if message.Level == LevelWarn {
			return true
		}
	}
	return false
}

func Evaluate(cfg config.Config, repo git.Repository) Report {
	var report Report

	changedFiles := repo.ChangedFiles()
	if cfg.Rules.MaxChangedFiles.Value > 0 && len(changedFiles) > cfg.Rules.MaxChangedFiles.Value {
		report.add(ruleLevel(cfg, cfg.Rules.MaxChangedFiles.Level, LevelFail), fmt.Sprintf("PR changes %d files, above the configured limit of %d", len(changedFiles), cfg.Rules.MaxChangedFiles.Value))
	}

	changedLines := repo.ChangedLines()
	if cfg.Rules.MaxChangedLines.Value > 0 && changedLines > cfg.Rules.MaxChangedLines.Value {
		report.add(ruleLevel(cfg, cfg.Rules.MaxChangedLines.Level, LevelFail), fmt.Sprintf("PR changes %d lines, above the configured limit of %d", changedLines, cfg.Rules.MaxChangedLines.Value))
	}

	if cfg.Rules.RequirePRTitlePattern.Value != "" {
		level := ruleLevel(cfg, cfg.Rules.RequirePRTitlePattern.Level, LevelFail)
		matched, err := regexp.MatchString(cfg.Rules.RequirePRTitlePattern.Value, repo.PullRequestTitle)
		if err != nil {
			report.fail(fmt.Sprintf("invalid require_pr_title_pattern: %s", err))
		} else if !matched {
			report.add(level, fmt.Sprintf("PR title %q does not match %q", repo.PullRequestTitle, cfg.Rules.RequirePRTitlePattern.Value))
		}
	}

	if cfg.Rules.RequireLinkedIssuePattern.Value != "" {
		level := ruleLevel(cfg, cfg.Rules.RequireLinkedIssuePattern.Level, LevelFail)
		text := strings.Join([]string{repo.PullRequestTitle, repo.PullRequestBody, repo.PullRequestBranch}, "\n")
		matched, err := regexp.MatchString(cfg.Rules.RequireLinkedIssuePattern.Value, text)
		if err != nil {
			report.fail(fmt.Sprintf("invalid require_linked_issue_pattern: %s", err))
		} else if !matched {
			report.add(level, fmt.Sprintf("PR title, body, or branch does not match linked issue pattern %q", cfg.Rules.RequireLinkedIssuePattern.Value))
		}
	}

	if cfg.Rules.RequireConventionalCommits.Enabled {
		level := ruleLevel(cfg, cfg.Rules.RequireConventionalCommits.Level, LevelFail)
		for _, commit := range repo.Commits {
			if !conventionalCommitPattern.MatchString(commit.Subject) {
				report.add(level, fmt.Sprintf("commit subject %q is not conventional", commit.Subject))
			}
		}
	}

	if cfg.Rules.RequireSignedOffCommits.Enabled {
		level := ruleLevel(cfg, cfg.Rules.RequireSignedOffCommits.Level, LevelFail)
		for _, commit := range repo.Commits {
			if !signedOffByPattern.MatchString(commit.Message) {
				report.add(level, fmt.Sprintf("commit %q is missing a Signed-off-by trailer; create commits with git commit --signoff", commit.Subject))
			}
		}
	}

	if cfg.Rules.RequireSquashedCommits.Enabled {
		level := ruleLevel(cfg, cfg.Rules.RequireSquashedCommits.Level, LevelFail)
		if len(repo.Commits) > 1 {
			report.add(level, fmt.Sprintf("PR has %d commits; squash to a single commit before merging", len(repo.Commits)))
		}
	}

	for _, label := range cfg.Rules.RequiredLabels.Values {
		if !hasLabel(repo.PullRequestLabels, label) {
			report.add(ruleLevel(cfg, cfg.Rules.RequiredLabels.Level, LevelFail), fmt.Sprintf("required label is missing: %s", label))
		}
	}

	for _, required := range cfg.Rules.RequiredFiles.Values {
		if !repo.HasFile(required) {
			report.add(ruleLevel(cfg, cfg.Rules.RequiredFiles.Level, LevelFail), fmt.Sprintf("required file is missing: %s", required))
		}
	}

	for _, required := range cfg.Rules.RequiredChangedFiles.Values {
		if !hasChangedFile(changedFiles, required) {
			report.add(ruleLevel(cfg, cfg.Rules.RequiredChangedFiles.Level, LevelFail), fmt.Sprintf("required changed file pattern is missing: %s", required))
		}
	}

	for _, forbidden := range cfg.Rules.ForbiddenFiles.Values {
		for _, changed := range changedFiles {
			if matchedPath(forbidden, changed) {
				report.add(ruleLevel(cfg, cfg.Rules.ForbiddenFiles.Level, LevelFail), fmt.Sprintf("forbidden file changed: %s", changed))
			}
		}
	}

	for _, warning := range cfg.Rules.WarnFiles.Values {
		for _, changed := range changedFiles {
			if matchedPath(warning, changed) {
				report.add(ruleLevel(cfg, cfg.Rules.WarnFiles.Level, LevelWarn), fmt.Sprintf("watched file changed: %s", changed))
			}
		}
	}

	if cfg.Rules.WarnDependencyChanges.Enabled {
		for _, changed := range changedFiles {
			if matchesAny(dependencyManifestPatterns, changed) {
				report.add(ruleLevel(cfg, cfg.Rules.WarnDependencyChanges.Level, LevelWarn), fmt.Sprintf("dependency manifest changed: %s", changed))
			}
		}
	}

	return report
}

func ruleLevel(cfg config.Config, configured string, fallback Level) Level {
	switch configured {
	case string(LevelFail):
		return LevelFail
	case string(LevelWarn):
		return LevelWarn
	}
	switch cfg.Level {
	case string(LevelFail):
		return LevelFail
	case string(LevelWarn):
		return LevelWarn
	}
	return fallback
}

func (r *Report) add(level Level, text string) {
	r.Messages = append(r.Messages, Message{Level: level, Text: text})
}

func (r *Report) fail(text string) {
	r.Messages = append(r.Messages, Message{Level: LevelFail, Text: text})
}

func (r *Report) warn(text string) {
	r.Messages = append(r.Messages, Message{Level: LevelWarn, Text: text})
}

func hasChangedFile(changedFiles []string, pattern string) bool {
	for _, changed := range changedFiles {
		if matchedPath(pattern, changed) {
			return true
		}
	}
	return false
}

func matchesAny(patterns []string, changed string) bool {
	for _, pattern := range patterns {
		if matchedPath(pattern, changed) {
			return true
		}
	}
	return false
}

func hasLabel(labels []string, required string) bool {
	for _, label := range labels {
		if strings.EqualFold(label, required) {
			return true
		}
	}
	return false
}

func matchedPath(pattern, path string) bool {
	normalizedPattern := normalizePath(pattern)
	normalizedPath := normalizePath(path)

	if normalizedPattern == normalizedPath {
		return true
	}

	if strings.HasSuffix(normalizedPattern, "/**") {
		prefix := strings.TrimSuffix(normalizedPattern, "/**")
		return normalizedPath == prefix || strings.HasPrefix(normalizedPath, prefix+"/")
	}

	matched, err := pathpkgMatch(normalizedPattern, normalizedPath)
	return err == nil && matched
}

func normalizePath(value string) string {
	return strings.TrimPrefix(strings.ReplaceAll(value, "\\", "/"), "./")
}

func pathpkgMatch(pattern, name string) (bool, error) {
	return path.Match(pattern, name)
}
