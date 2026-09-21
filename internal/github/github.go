package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/harryvince/danger-go/internal/danger"
	"github.com/harryvince/danger-go/internal/git"
)

const apiBase = "https://api.github.com"
const commentMarker = "<!-- danger-go:report -->"

const (
	LabelPassed = "danger::passed"
	LabelWarn   = "danger::warn"
	LabelFail   = "danger::fail"
)

var statusLabels = []string{LabelPassed, LabelWarn, LabelFail}

var statusLabelColors = map[string]string{
	LabelPassed: "2da44e",
	LabelWarn:   "bf8700",
	LabelFail:   "cf222e",
}

type Client struct {
	token      string
	httpClient *http.Client
	apiBase    string
}

type PullRequestContext struct {
	Owner  string
	Repo   string
	Number int
	Title  string
	Body   string
	Branch string
	Labels []string
}

type actionsEvent struct {
	PullRequest struct {
		Number int    `json:"number"`
		Title  string `json:"title"`
		Body   string `json:"body"`
		Head   struct {
			Ref string `json:"ref"`
		} `json:"head"`
		Labels []struct {
			Name string `json:"name"`
		} `json:"labels"`
	} `json:"pull_request"`
	Repository struct {
		Name  string `json:"name"`
		Owner struct {
			Login string `json:"login"`
		} `json:"owner"`
	} `json:"repository"`
}

type fileResponse struct {
	Filename  string `json:"filename"`
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions"`
}

type commitResponse struct {
	SHA    string `json:"sha"`
	Commit struct {
		Message string `json:"message"`
	} `json:"commit"`
}

type commentResponse struct {
	ID   int64  `json:"id"`
	Body string `json:"body"`
	User struct {
		Type string `json:"type"`
	} `json:"user"`
}

func NewClientFromEnv() *Client {
	return &Client{
		token:      firstNonEmpty(os.Getenv("GITHUB_TOKEN"), os.Getenv("GH_TOKEN")),
		httpClient: http.DefaultClient,
		apiBase:    apiBase,
	}
}

func InActions() bool {
	return os.Getenv("GITHUB_ACTIONS") == "true"
}

func ContextFromActions() (*PullRequestContext, error) {
	eventName := os.Getenv("GITHUB_EVENT_NAME")
	if eventName != "pull_request" && eventName != "pull_request_target" {
		return nil, fmt.Errorf("danger-go ci currently supports pull_request events, got %q", eventName)
	}

	eventPath := os.Getenv("GITHUB_EVENT_PATH")
	if eventPath == "" {
		return nil, fmt.Errorf("GITHUB_EVENT_PATH is not set")
	}

	data, err := os.ReadFile(eventPath)
	if err != nil {
		return nil, err
	}

	var event actionsEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, err
	}

	if event.PullRequest.Number == 0 {
		return nil, fmt.Errorf("GitHub event payload does not contain a pull request")
	}

	return &PullRequestContext{
		Owner:  event.Repository.Owner.Login,
		Repo:   event.Repository.Name,
		Number: event.PullRequest.Number,
		Title:  event.PullRequest.Title,
		Body:   event.PullRequest.Body,
		Branch: event.PullRequest.Head.Ref,
		Labels: eventLabels(event),
	}, nil
}

func (c *Client) Repository(ctx context.Context, pr PullRequestContext) (git.Repository, error) {
	files, err := c.changedFiles(ctx, pr)
	if err != nil {
		return git.Repository{}, err
	}

	commits, err := c.commits(ctx, pr)
	if err != nil {
		return git.Repository{}, err
	}

	return git.Repository{
		ModifiedFiles:     changePaths(files),
		FileChanges:       fileChanges(files),
		Commits:           gitCommits(commits),
		PullRequestTitle:  pr.Title,
		PullRequestBody:   pr.Body,
		PullRequestBranch: pr.Branch,
		PullRequestLabels: pr.Labels,
	}, nil
}

func eventLabels(event actionsEvent) []string {
	labels := make([]string, 0, len(event.PullRequest.Labels))
	for _, label := range event.PullRequest.Labels {
		if label.Name != "" {
			labels = append(labels, label.Name)
		}
	}
	return labels
}

func (c *Client) SyncReportLabel(ctx context.Context, pr PullRequestContext, report danger.Report) error {
	target := reportLabel(report)
	if err := c.ensureLabel(ctx, pr, target); err != nil {
		return err
	}

	for _, label := range statusLabels {
		if label == target {
			continue
		}
		if err := c.removeIssueLabel(ctx, pr, label); err != nil {
			return err
		}
	}

	return c.addIssueLabel(ctx, pr, target)
}

func (c *Client) PostReportComment(ctx context.Context, pr PullRequestContext, report danger.Report) error {
	body := renderComment(report)
	payload, err := json.Marshal(map[string]string{"body": body})
	if err != nil {
		return err
	}

	existing, err := c.findReportComment(ctx, pr)
	if err != nil {
		return err
	}

	method := http.MethodPost
	path := fmt.Sprintf("/repos/%s/%s/issues/%d/comments", pr.Owner, pr.Repo, pr.Number)
	if existing != nil {
		method = http.MethodPatch
		path = fmt.Sprintf("/repos/%s/%s/issues/comments/%d", pr.Owner, pr.Repo, existing.ID)
	}

	req, err := c.request(ctx, method, path, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("posting GitHub comment failed: %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	return nil
}

func (c *Client) ensureLabel(ctx context.Context, pr PullRequestContext, label string) error {
	payload, err := json.Marshal(map[string]string{
		"name":        label,
		"color":       statusLabelColors[label],
		"description": "danger-go status",
	})
	if err != nil {
		return err
	}

	req, err := c.request(ctx, http.MethodPost, fmt.Sprintf("/repos/%s/%s/labels", pr.Owner, pr.Repo), bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusUnprocessableEntity {
		return nil
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("creating GitHub label failed: %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	return nil
}

func (c *Client) addIssueLabel(ctx context.Context, pr PullRequestContext, label string) error {
	payload, err := json.Marshal(map[string][]string{"labels": []string{label}})
	if err != nil {
		return err
	}

	req, err := c.request(ctx, http.MethodPost, fmt.Sprintf("/repos/%s/%s/issues/%d/labels", pr.Owner, pr.Repo, pr.Number), bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("adding GitHub label failed: %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	return nil
}

func (c *Client) removeIssueLabel(ctx context.Context, pr PullRequestContext, label string) error {
	path := fmt.Sprintf("/repos/%s/%s/issues/%d/labels/%s", pr.Owner, pr.Repo, pr.Number, url.PathEscape(label))
	req, err := c.request(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("removing GitHub label failed: %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	return nil
}

func reportLabel(report danger.Report) string {
	if report.HasFailures() {
		return LabelFail
	}
	if report.HasWarnings() {
		return LabelWarn
	}
	return LabelPassed
}

func (c *Client) findReportComment(ctx context.Context, pr PullRequestContext) (*commentResponse, error) {
	for page := 1; ; page++ {
		path := fmt.Sprintf("/repos/%s/%s/issues/%d/comments?per_page=100&page=%d", pr.Owner, pr.Repo, pr.Number, page)
		req, err := c.request(ctx, http.MethodGet, path, nil)
		if err != nil {
			return nil, err
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, err
		}

		var comments []commentResponse
		if err := decodeResponse(resp, &comments); err != nil {
			return nil, err
		}

		for _, comment := range comments {
			if strings.Contains(comment.Body, commentMarker) && comment.User.Type == "Bot" {
				return &comment, nil
			}
		}

		if len(comments) < 100 {
			return nil, nil
		}
	}
}

func (c *Client) changedFiles(ctx context.Context, pr PullRequestContext) ([]fileResponse, error) {
	var files []fileResponse
	for page := 1; ; page++ {
		path := fmt.Sprintf("/repos/%s/%s/pulls/%d/files?per_page=100&page=%d", pr.Owner, pr.Repo, pr.Number, page)
		req, err := c.request(ctx, http.MethodGet, path, nil)
		if err != nil {
			return nil, err
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, err
		}

		var pageFiles []fileResponse
		if err := decodeResponse(resp, &pageFiles); err != nil {
			return nil, err
		}
		if len(pageFiles) == 0 {
			break
		}

		files = append(files, pageFiles...)
		if len(pageFiles) < 100 {
			break
		}
	}
	return files, nil
}

func (c *Client) commits(ctx context.Context, pr PullRequestContext) ([]commitResponse, error) {
	var commits []commitResponse
	for page := 1; ; page++ {
		path := fmt.Sprintf("/repos/%s/%s/pulls/%d/commits?per_page=100&page=%d", pr.Owner, pr.Repo, pr.Number, page)
		req, err := c.request(ctx, http.MethodGet, path, nil)
		if err != nil {
			return nil, err
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, err
		}

		var pageCommits []commitResponse
		if err := decodeResponse(resp, &pageCommits); err != nil {
			return nil, err
		}
		if len(pageCommits) == 0 {
			break
		}

		commits = append(commits, pageCommits...)
		if len(pageCommits) < 100 {
			break
		}
	}
	return commits, nil
}

func changePaths(files []fileResponse) []string {
	paths := make([]string, 0, len(files))
	for _, file := range files {
		paths = append(paths, file.Filename)
	}
	return paths
}

func fileChanges(files []fileResponse) []git.FileChange {
	changes := make([]git.FileChange, 0, len(files))
	for _, file := range files {
		changes = append(changes, git.FileChange{
			Path:      file.Filename,
			Additions: file.Additions,
			Deletions: file.Deletions,
		})
	}
	return changes
}

func gitCommits(commits []commitResponse) []git.Commit {
	result := make([]git.Commit, 0, len(commits))
	for _, commit := range commits {
		subject := strings.SplitN(commit.Commit.Message, "\n", 2)[0]
		if subject == "" {
			continue
		}
		result = append(result, git.Commit{SHA: commit.SHA, Subject: subject})
	}
	return result
}

func (c *Client) request(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	base := c.apiBase
	if base == "" {
		base = apiBase
	}
	req, err := http.NewRequestWithContext(ctx, method, base+path, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	return req, nil
}

func decodeResponse(resp *http.Response, target any) error {
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GitHub API request failed: %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	return json.NewDecoder(resp.Body).Decode(target)
}

func renderComment(report danger.Report) string {
	var b strings.Builder
	b.WriteString(commentMarker)
	b.WriteString("\n")
	b.WriteString("## danger-go\n\n")
	if len(report.Messages) == 0 {
		b.WriteString("No issues found.\n")
		return b.String()
	}
	for _, message := range report.Messages {
		b.WriteString(fmt.Sprintf("- **%s**: %s\n", message.Level, message.Text))
	}
	return b.String()
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
