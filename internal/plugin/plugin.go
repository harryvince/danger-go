package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"

	"github.com/harryvince/danger-go/internal/config"
	"github.com/harryvince/danger-go/internal/danger"
	"github.com/harryvince/danger-go/internal/git"
)

const ProtocolVersion = "v1"

type Request struct {
	ProtocolVersion string         `json:"protocol_version"`
	Repo            RepositoryInfo `json:"repo"`
	Config          map[string]any `json:"config,omitempty"`
}

type RepositoryInfo struct {
	Files        []string         `json:"files"`
	Changed      []git.FileChange `json:"changed"`
	ChangedFiles []string         `json:"changed_files"`
	ChangedLines int              `json:"changed_lines"`
	PullRequest  PullRequestInfo  `json:"pull_request"`
	Commits      []git.Commit     `json:"commits"`
}

type PullRequestInfo struct {
	Title  string   `json:"title"`
	Body   string   `json:"body"`
	Branch string   `json:"branch"`
	Labels []string `json:"labels"`
}

type Response struct {
	Messages []Message `json:"messages"`
}

type Message struct {
	Level string `json:"level"`
	Text  string `json:"text"`
}

func RunAll(ctx context.Context, plugins []config.Plugin, repo git.Repository, defaultLevel string) danger.Report {
	var report danger.Report
	for _, p := range plugins {
		for _, msg := range Run(ctx, p, repo, defaultLevel).Messages {
			report.Messages = append(report.Messages, msg)
		}
	}
	return report
}

func Run(ctx context.Context, p config.Plugin, repo git.Repository, defaultLevel string) danger.Report {
	if len(p.Command) == 0 {
		return failure(p.Name, "plugin command is empty")
	}

	timeout := 30 * time.Second
	if p.Timeout != "" {
		parsed, err := time.ParseDuration(p.Timeout)
		if err != nil {
			return failure(p.Name, fmt.Sprintf("invalid plugin timeout %q: %s", p.Timeout, err))
		}
		timeout = parsed
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	body, err := json.Marshal(Request{ProtocolVersion: ProtocolVersion, Repo: repoInfo(repo), Config: p.Config})
	if err != nil {
		return failure(p.Name, fmt.Sprintf("could not encode plugin request: %s", err))
	}

	cmd := exec.CommandContext(ctx, p.Command[0], p.Command[1:]...)
	cmd.Stdin = bytes.NewReader(body)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return failure(p.Name, fmt.Sprintf("plugin timed out after %s", timeout))
		}
		return failure(p.Name, fmt.Sprintf("plugin exited with error: %s%s", err, stderrSuffix(stderr.String())))
	}

	var response Response
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		return failure(p.Name, fmt.Sprintf("plugin returned invalid JSON: %s", err))
	}

	var report danger.Report
	for _, message := range response.Messages {
		level := pluginLevel(message.Level, p.Level, defaultLevel)
		report.Messages = append(report.Messages, danger.Message{Level: level, Text: fmt.Sprintf("[%s] %s", p.Name, message.Text)})
	}
	return report
}

func repoInfo(repo git.Repository) RepositoryInfo {
	return RepositoryInfo{
		Files:        repo.Files,
		Changed:      repo.FileChanges,
		ChangedFiles: repo.ChangedFiles(),
		ChangedLines: repo.ChangedLines(),
		PullRequest:  PullRequestInfo{Title: repo.PullRequestTitle, Body: repo.PullRequestBody, Branch: repo.PullRequestBranch, Labels: repo.PullRequestLabels},
		Commits:      repo.Commits,
	}
}

func pluginLevel(messageLevel, configured, defaultLevel string) danger.Level {
	for _, value := range []string{messageLevel, configured, defaultLevel} {
		switch value {
		case string(danger.LevelWarn):
			return danger.LevelWarn
		case string(danger.LevelFail):
			return danger.LevelFail
		}
	}
	return danger.LevelFail
}

func failure(name, text string) danger.Report {
	return danger.Report{Messages: []danger.Message{{Level: danger.LevelFail, Text: fmt.Sprintf("[%s] %s", name, text)}}}
}

func stderrSuffix(stderr string) string {
	if stderr == "" {
		return ""
	}
	return ": " + stderr
}
