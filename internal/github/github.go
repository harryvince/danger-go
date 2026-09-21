package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/harryvince/danger-go/internal/danger"
	"github.com/harryvince/danger-go/internal/git"
)

const apiBase = "https://api.github.com"
const commentMarker = "<!-- danger-go:report -->"

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
}

type actionsEvent struct {
	PullRequest struct {
		Number int    `json:"number"`
		Title  string `json:"title"`
	} `json:"pull_request"`
	Repository struct {
		Name  string `json:"name"`
		Owner struct {
			Login string `json:"login"`
		} `json:"owner"`
	} `json:"repository"`
}

type fileResponse struct {
	Filename string `json:"filename"`
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
	}, nil
}

func (c *Client) Repository(ctx context.Context, pr PullRequestContext) (git.Repository, error) {
	files, err := c.changedFiles(ctx, pr)
	if err != nil {
		return git.Repository{}, err
	}

	return git.Repository{
		ModifiedFiles:    files,
		PullRequestTitle: pr.Title,
	}, nil
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

func (c *Client) changedFiles(ctx context.Context, pr PullRequestContext) ([]string, error) {
	var files []string
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

		for _, file := range pageFiles {
			files = append(files, file.Filename)
		}
		if len(pageFiles) < 100 {
			break
		}
	}
	return files, nil
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
