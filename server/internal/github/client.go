// Package github is a deliberately minimal GitHub REST client scoped
// to the single call the server makes today: creating an issue in the
// project repository (the in-app "report a bug" button). It is NOT a
// general GitHub SDK — no pagination, no retries, no rate-limit
// bookkeeping. If the server ever needs a second endpoint, grow this
// package call-by-call rather than importing a third-party client
// (same outbound-HTTP posture as the S06.5 deck importers: stdlib
// only, explicit timeout, descriptive User-Agent).
//
// The token is a fine-grained PAT with Issues:write on the one repo
// (see ADR 0017). It is held in memory only; it never appears in
// logs, error strings, or responses.
package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	defaultBaseURL = "https://api.github.com"

	// defaultTimeout matches the S06.5 outbound-HTTP hygiene rule:
	// no outbound call may hang a handler for more than ~10s.
	defaultTimeout = 10 * time.Second

	// maxResponseBytes caps how much of a GitHub response we will
	// read (same defensive posture as the avatar cache's size cap).
	// Issue-creation responses are ~2 KB; 1 MiB is generous.
	maxResponseBytes = 1 << 20

	userAgent = "cmd_and_ctrl-server (+https://github.com/krakenhavoc/cmd_and_ctrl)"
)

// APIError is a non-2xx response from the GitHub API. It exists so a
// caller can tell "GitHub rejected this request" (a 4xx — the issue
// was definitively NOT created) from "the call didn't complete" (a
// timeout, a 5xx), without parsing the error string.
//
// The lobby uses it for exactly one decision: an issue whose labels
// GitHub refuses is re-filed unlabelled rather than lost (ADR 0017
// §8). It satisfies the unexported statusCoder interface lobby
// declares, so the BugReporter seam stays one method wide and this
// package stays un-imported there.
type APIError struct {
	Status int
	Body   string // already truncated
}

func (e *APIError) Error() string {
	return fmt.Sprintf("github: create issue: status %d: %s", e.Status, e.Body)
}

// StatusCode reports the HTTP status GitHub answered with.
func (e *APIError) StatusCode() int { return e.Status }

// Client files issues against a single owner/name repository.
type Client struct {
	baseURL string
	repo    string // "owner/name"
	token   string
	http    *http.Client
}

// NewClient returns a Client for the given fine-grained token and
// "owner/name" repo slug.
func NewClient(token, repo string) *Client {
	return &Client{
		baseURL: defaultBaseURL,
		repo:    strings.Trim(repo, "/"),
		token:   token,
		http:    &http.Client{Timeout: defaultTimeout},
	}
}

// WithHTTPClient injects a custom http.Client. Intended for tests
// that drive against an httptest.Server.
func (c *Client) WithHTTPClient(h *http.Client) *Client {
	c.http = h
	return c
}

// WithBaseURL overrides the API base URL. Intended for tests.
func (c *Client) WithBaseURL(u string) *Client {
	c.baseURL = strings.TrimRight(u, "/")
	return c
}

// CreateIssue files an issue and returns its html_url and number.
// The signature deliberately matches lobby.BugReporter so main.go
// can wire the concrete client straight into lobby.Config without an
// adapter.
//
// Unknown labels are not pre-validated: GitHub attaches existing
// ones and (for tokens with push access) creates missing ones; a
// label quirk must never block the report itself. When GitHub does
// reject the labels it answers 4xx, and the error is an *APIError
// carrying that status so the caller can re-file unlabelled rather
// than drop the report (ADR 0017 §8).
func (c *Client) CreateIssue(ctx context.Context, title, body string, labels []string) (string, int, error) {
	payload := struct {
		Title  string   `json:"title"`
		Body   string   `json:"body"`
		Labels []string `json:"labels,omitempty"`
	}{Title: title, Body: body, Labels: labels}

	buf, err := json.Marshal(payload)
	if err != nil {
		return "", 0, fmt.Errorf("github: encode issue: %w", err)
	}

	url := fmt.Sprintf("%s/repos/%s/issues", c.baseURL, c.repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(buf))
	if err != nil {
		return "", 0, fmt.Errorf("github: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.http.Do(req)
	if err != nil {
		// err carries the URL but never the Authorization header,
		// so it is safe to propagate verbatim.
		return "", 0, fmt.Errorf("github: create issue: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	lim := io.LimitReader(resp.Body, maxResponseBytes)
	if resp.StatusCode != http.StatusCreated {
		slurp, _ := io.ReadAll(lim)
		return "", 0, &APIError{Status: resp.StatusCode, Body: truncate(string(slurp), 300)}
	}

	var out struct {
		Number  int    `json:"number"`
		HTMLURL string `json:"html_url"`
	}
	if err := json.NewDecoder(lim).Decode(&out); err != nil {
		return "", 0, fmt.Errorf("github: decode response: %w", err)
	}
	return out.HTMLURL, out.Number, nil
}

// truncate clips s to at most n bytes, appending an ellipsis marker
// when it clipped. GitHub error bodies are JSON blobs that can run
// long; error strings should stay log-line sized.
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…(truncated)"
}
