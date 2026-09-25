package github

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Tests for the two calls ADR 0095's deck requests added: reading an
// issue's state and commenting on it.

func TestGetIssueReadsState(t *testing.T) {
	var gotPath, gotAuth, gotCT string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.Method + " " + r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		gotCT = r.Header.Get("Content-Type")
		_, _ = w.Write([]byte(`{"number": 7, "state": "closed", "html_url": "https://github.com/o/r/issues/7", "title": "x"}`))
	}))
	defer srv.Close()

	c := NewClient("tok-abc", "o/r").WithBaseURL(srv.URL).WithHTTPClient(srv.Client())
	iss, err := c.GetIssue(context.Background(), 7)
	if err != nil {
		t.Fatalf("GetIssue: %v", err)
	}
	if iss.Number != 7 || iss.State != "closed" || iss.Open() || iss.HTMLURL != "https://github.com/o/r/issues/7" {
		t.Errorf("issue = %+v", iss)
	}
	if gotPath != "GET /repos/o/r/issues/7" {
		t.Errorf("path = %q", gotPath)
	}
	if gotAuth != "Bearer tok-abc" {
		t.Errorf("auth = %q", gotAuth)
	}
	if gotCT != "" {
		t.Errorf("a GET should carry no Content-Type, got %q", gotCT)
	}
}

func TestGetIssueGoneCarriesStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusGone)
		_, _ = w.Write([]byte(`{"message":"This issue was deleted"}`))
	}))
	defer srv.Close()

	c := NewClient("tok", "o/r").WithBaseURL(srv.URL).WithHTTPClient(srv.Client())
	_, err := c.GetIssue(context.Background(), 9)
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode() != http.StatusGone {
		t.Fatalf("err = %v, want an *APIError with 410", err)
	}
	if !strings.Contains(err.Error(), "get issue") {
		t.Errorf("error should name the call: %v", err)
	}
}

func TestCreateCommentPostsBody(t *testing.T) {
	var gotPath string
	var got struct {
		Body string `json:"body"`
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.Method + " " + r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&got)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id": 1, "html_url": "https://github.com/o/r/issues/7#issuecomment-1"}`))
	}))
	defer srv.Close()

	c := NewClient("tok", "o/r").WithBaseURL(srv.URL).WithHTTPClient(srv.Client())
	url, err := c.CreateComment(context.Background(), 7, "Also requested by Alice")
	if err != nil {
		t.Fatalf("CreateComment: %v", err)
	}
	if url != "https://github.com/o/r/issues/7#issuecomment-1" {
		t.Errorf("url = %q", url)
	}
	if gotPath != "POST /repos/o/r/issues/7/comments" || got.Body != "Also requested by Alice" {
		t.Errorf("sent %q with %+v", gotPath, got)
	}
}

func TestCreateCommentErrorNamesTheCall(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	c := NewClient("tok", "o/r").WithBaseURL(srv.URL).WithHTTPClient(srv.Client())
	_, err := c.CreateComment(context.Background(), 7, "x")
	if err == nil || !strings.Contains(err.Error(), "create comment: status 403") {
		t.Fatalf("err = %v", err)
	}
	if strings.Contains(err.Error(), "tok") {
		t.Errorf("error must not leak the token: %v", err)
	}
}
