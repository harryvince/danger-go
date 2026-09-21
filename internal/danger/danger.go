package danger

import (
	"fmt"
	"path/filepath"
	"regexp"

	"github.com/harryvince/danger-go/internal/config"
	"github.com/harryvince/danger-go/internal/git"
)

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

func Evaluate(cfg config.Config, repo git.Repository) Report {
	var report Report

	changedFiles := repo.ChangedFiles()
	if cfg.Rules.MaxChangedFiles > 0 && len(changedFiles) > cfg.Rules.MaxChangedFiles {
		report.fail(fmt.Sprintf("PR changes %d files, above the configured limit of %d", len(changedFiles), cfg.Rules.MaxChangedFiles))
	}

	if cfg.Rules.RequirePRTitlePattern != "" {
		matched, err := regexp.MatchString(cfg.Rules.RequirePRTitlePattern, repo.PullRequestTitle)
		if err != nil {
			report.fail(fmt.Sprintf("invalid require_pr_title_pattern: %s", err))
		} else if !matched {
			report.fail(fmt.Sprintf("PR title %q does not match %q", repo.PullRequestTitle, cfg.Rules.RequirePRTitlePattern))
		}
	}

	for _, required := range cfg.Rules.RequiredFiles {
		if !repo.HasFile(required) {
			report.fail(fmt.Sprintf("required file is missing: %s", required))
		}
	}

	for _, forbidden := range cfg.Rules.ForbiddenFiles {
		for _, changed := range changedFiles {
			if matchedPath(forbidden, changed) {
				report.fail(fmt.Sprintf("forbidden file changed: %s", changed))
			}
		}
	}

	return report
}

func (r *Report) fail(text string) {
	r.Messages = append(r.Messages, Message{Level: LevelFail, Text: text})
}

func matchedPath(pattern, path string) bool {
	if pattern == path {
		return true
	}

	matched, err := filepath.Match(pattern, path)
	return err == nil && matched
}
