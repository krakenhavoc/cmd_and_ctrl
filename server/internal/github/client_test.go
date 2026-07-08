package github

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCreateIssueHappyPath(t *testing.T) {
	var gotPath, gotAuth, gotUA, gotAccept string
	var gotPayload struct {
		Title  string   `json:"title"`
		Body   string   `json:"body"`
		Labels []string `json:"labels"`
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.Method + " " + r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		gotUA = r.Header.Get("User-Agent")
		gotAccept = r.Header.Get("Accept")
		if err := json.NewDecoder(r.Body).Decode(&gotPayload); err != nil {
			t.Errorf("decode payload: %v", err)
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"number": 42, "html_url": "https://github.com/o/r/issues/42"}`))
	}))
	defer srv.Close()

	c := NewClient("tok-abc", "o/r").WithBaseURL(srv.URL).WithHTTPClient(srv.Client())
	url, number, err := c.CreateIssue(context.Background(), "[in-app] it broke", "body text", []string{"bug"})
	if err != nil {
		t.Fatalf("CreateIssue: %v", err)
	}
	if url != "https://github.com/o/r/issues/42" || number != 42 {
		t.Fatalf("got url=%q number=%d", url, number)
	}
	if gotPath != "POST /repos/o/r/issues" {
		t.Errorf("path = %q", gotPath)
	}
	if gotAuth != "Bearer tok-abc" {
		t.Errorf("auth header = %q", gotAuth)
	}
	if !strings.Contains(gotUA, "cmd_and_ctrl") {
		t.Errorf("user-agent = %q", gotUA)
	}
	if gotAccept != "application/vnd.github+json" {
		t.Errorf("accept = %q", gotAccept)
	}
	if gotPayload.Title != "[in-app] it broke" || gotPayload.Body != "body text" {
		t.Errorf("payload = %+v", gotPayload)
	}
	if len(gotPayload.Labels) != 1 || gotPayload.Labels[0] != "bug" {
		t.Errorf("labels = %v", gotPayload.Labels)
	}
}

func TestCreateIssueNon201IsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"message":"Validation Failed"}`))
	}))
	defer srv.Close()

	c := NewClient("tok", "o/r").WithBaseURL(srv.URL).WithHTTPClient(srv.Client())
	_, _, err := c.CreateIssue(context.Background(), "t", "b", nil)
	if err == nil {
		t.Fatal("expected error on 422")
	}
	if !strings.Contains(err.Error(), "422") {
		t.Errorf("error should carry status: %v", err)
	}
	if strings.Contains(err.Error(), "tok") {
		t.Errorf("error must not leak the token: %v", err)
	}
}

func TestCreateIssueTruncatesHugeErrorBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(strings.Repeat("x", 10_000)))
	}))
	defer srv.Close()

	c := NewClient("tok", "o/r").WithBaseURL(srv.URL).WithHTTPClient(srv.Client())
	_, _, err := c.CreateIssue(context.Background(), "t", "b", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if len(err.Error()) > 500 {
		t.Errorf("error string not truncated: %d bytes", len(err.Error()))
	}
}
