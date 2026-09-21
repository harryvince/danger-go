package github

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/harryvince/danger-go/internal/danger"
)

func TestPostReportCommentCreatesWhenNoExistingComment(t *testing.T) {
	var methods []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		methods = append(methods, r.Method+" "+r.URL.Path)
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/repos/o/r/issues/7/comments":
			writeJSON(t, w, []commentResponse{})
		case r.Method == http.MethodPost && r.URL.Path == "/repos/o/r/issues/7/comments":
			var payload map[string]string
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(payload["body"], commentMarker) {
				t.Fatalf("comment body missing marker: %q", payload["body"])
			}
			w.WriteHeader(http.StatusCreated)
			writeJSON(t, w, map[string]int{"id": 123})
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
		}
	}))
	defer server.Close()

	client := testClient(server.URL)
	err := client.PostReportComment(context.Background(), PullRequestContext{Owner: "o", Repo: "r", Number: 7}, danger.Report{})
	if err != nil {
		t.Fatal(err)
	}

	if got, want := strings.Join(methods, ","), "GET /repos/o/r/issues/7/comments,POST /repos/o/r/issues/7/comments"; got != want {
		t.Fatalf("methods = %s, want %s", got, want)
	}
}

func TestContextFromActionsReadsPullRequestMetadata(t *testing.T) {
	t.Setenv("GITHUB_EVENT_NAME", "pull_request")
	eventPath := filepath.Join(t.TempDir(), "event.json")
	event := `{
	  "pull_request": {
	    "number": 42,
	    "title": "JIRA-123: add rule",
	    "body": "Closes JIRA-123",
	    "head": {"ref": "feature/JIRA-123-rule"},
	    "labels": [{"name": "ready"}, {"name": "rules"}]
	  },
	  "repository": {
	    "name": "danger-go",
	    "owner": {"login": "harryvince"}
	  }
	}`
	if err := os.WriteFile(eventPath, []byte(event), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GITHUB_EVENT_PATH", eventPath)

	ctx, err := ContextFromActions()
	if err != nil {
		t.Fatal(err)
	}

	if ctx.Body != "Closes JIRA-123" {
		t.Fatalf("body = %q", ctx.Body)
	}
	if ctx.Branch != "feature/JIRA-123-rule" {
		t.Fatalf("branch = %q", ctx.Branch)
	}
	if got, want := strings.Join(ctx.Labels, ","), "ready,rules"; got != want {
		t.Fatalf("labels = %q, want %q", got, want)
	}
}

func TestPostReportCommentUpdatesExistingComment(t *testing.T) {
	var methods []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		methods = append(methods, r.Method+" "+r.URL.Path)
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/repos/o/r/issues/7/comments":
			writeJSON(t, w, []commentResponse{{
				ID:   456,
				Body: commentMarker + "\n## danger-go\n\nOld result\n",
				User: struct {
					Type string `json:"type"`
				}{Type: "Bot"},
			}})
		case r.Method == http.MethodPatch && r.URL.Path == "/repos/o/r/issues/comments/456":
			w.WriteHeader(http.StatusOK)
			writeJSON(t, w, map[string]int{"id": 456})
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
		}
	}))
	defer server.Close()

	client := testClient(server.URL)
	err := client.PostReportComment(context.Background(), PullRequestContext{Owner: "o", Repo: "r", Number: 7}, danger.Report{})
	if err != nil {
		t.Fatal(err)
	}

	if got, want := strings.Join(methods, ","), "GET /repos/o/r/issues/7/comments,PATCH /repos/o/r/issues/comments/456"; got != want {
		t.Fatalf("methods = %s, want %s", got, want)
	}
}

func testClient(apiBase string) *Client {
	return &Client{
		httpClient: http.DefaultClient,
		apiBase:    apiBase,
	}
}

func writeJSON(t *testing.T, w http.ResponseWriter, value any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(value); err != nil {
		t.Fatal(err)
	}
}
