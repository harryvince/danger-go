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
	    "title": "ISSUE-123: add rule",
	    "body": "Closes ISSUE-123",
	    "head": {"ref": "feature/ISSUE-123-rule"},
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

	if ctx.Body != "Closes ISSUE-123" {
		t.Fatalf("body = %q", ctx.Body)
	}
	if ctx.Branch != "feature/ISSUE-123-rule" {
		t.Fatalf("branch = %q", ctx.Branch)
	}
	if got, want := strings.Join(ctx.Labels, ","), "ready,rules"; got != want {
		t.Fatalf("labels = %q, want %q", got, want)
	}
}

func TestRepositoryReadsChangedFilesAndCommits(t *testing.T) {
	var methods []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		methods = append(methods, r.Method+" "+r.URL.String())
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/repos/o/r/pulls/7/files":
			writeJSON(t, w, []fileResponse{{
				Filename:  "README.md",
				Additions: 5,
				Deletions: 1,
			}})
		case r.Method == http.MethodGet && r.URL.Path == "/repos/o/r/pulls/7/commits":
			writeJSON(t, w, []commitResponse{{
				SHA: "abc123",
				Commit: struct {
					Message string `json:"message"`
				}{Message: "feat: add commit policy\n\nBody"},
			}})
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
		}
	}))
	defer server.Close()

	client := testClient(server.URL)
	repo, err := client.Repository(context.Background(), PullRequestContext{Owner: "o", Repo: "r", Number: 7})
	if err != nil {
		t.Fatal(err)
	}

	if got, want := strings.Join(repo.ChangedFiles(), ","), "README.md"; got != want {
		t.Fatalf("changed files = %q, want %q", got, want)
	}
	if got, want := repo.ChangedLines(), 6; got != want {
		t.Fatalf("changed lines = %d, want %d", got, want)
	}
	if got, want := len(repo.Commits), 1; got != want {
		t.Fatalf("commit count = %d, want %d", got, want)
	}
	if repo.Commits[0].Subject != "feat: add commit policy" {
		t.Fatalf("commit subject = %q", repo.Commits[0].Subject)
	}

	want := "GET /repos/o/r/pulls/7/files?per_page=100&page=1,GET /repos/o/r/pulls/7/commits?per_page=100&page=1"
	if got := strings.Join(methods, ","); got != want {
		t.Fatalf("methods = %s, want %s", got, want)
	}
}

func TestSyncReportLabelAddsPassedAndRemovesStaleLabels(t *testing.T) {
	var methods []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		methods = append(methods, r.Method+" "+r.URL.Path)
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/repos/o/r/labels":
			var payload map[string]string
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatal(err)
			}
			if payload["name"] != LabelPassed {
				t.Fatalf("label name = %q, want %q", payload["name"], LabelPassed)
			}
			w.WriteHeader(http.StatusUnprocessableEntity)
			writeJSON(t, w, map[string]string{"message": "already exists"})
		case r.Method == http.MethodDelete && r.URL.Path == "/repos/o/r/issues/7/labels/danger::warn":
			w.WriteHeader(http.StatusNotFound)
			writeJSON(t, w, map[string]string{"message": "not found"})
		case r.Method == http.MethodDelete && r.URL.Path == "/repos/o/r/issues/7/labels/danger::fail":
			w.WriteHeader(http.StatusOK)
			writeJSON(t, w, []map[string]string{})
		case r.Method == http.MethodPost && r.URL.Path == "/repos/o/r/issues/7/labels":
			var payload map[string][]string
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatal(err)
			}
			if got := payload["labels"]; len(got) != 1 || got[0] != LabelPassed {
				t.Fatalf("labels payload = %#v", payload)
			}
			w.WriteHeader(http.StatusOK)
			writeJSON(t, w, []map[string]string{{"name": LabelPassed}})
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
		}
	}))
	defer server.Close()

	client := testClient(server.URL)
	if err := client.SyncReportLabel(context.Background(), PullRequestContext{Owner: "o", Repo: "r", Number: 7}, danger.Report{}); err != nil {
		t.Fatal(err)
	}

	want := "POST /repos/o/r/labels,DELETE /repos/o/r/issues/7/labels/danger::warn,DELETE /repos/o/r/issues/7/labels/danger::fail,POST /repos/o/r/issues/7/labels"
	if got := strings.Join(methods, ","); got != want {
		t.Fatalf("methods = %s, want %s", got, want)
	}
}

func TestReportLabelPrefersFailThenWarn(t *testing.T) {
	report := danger.Report{Messages: []danger.Message{
		{Level: danger.LevelWarn, Text: "warning"},
		{Level: danger.LevelFail, Text: "failure"},
	}}
	if got := reportLabel(report); got != LabelFail {
		t.Fatalf("label = %q, want %q", got, LabelFail)
	}

	report = danger.Report{Messages: []danger.Message{{Level: danger.LevelWarn, Text: "warning"}}}
	if got := reportLabel(report); got != LabelWarn {
		t.Fatalf("label = %q, want %q", got, LabelWarn)
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
