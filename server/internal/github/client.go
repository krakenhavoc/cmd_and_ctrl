// Package github is a deliberately minimal GitHub REST client scoped
// to the calls the server makes: creating an issue in the project
// repository (the in-app "report a bug" button), and, for ADR 0095's
// deck requests, reading an issue's state and commenting on it. It is
// NOT a general GitHub SDK — no pagination, no retries, no rate-limit
// bookkeeping. If the server ever needs another endpoint, grow this
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
// The lobby uses it for two decisions: an issue whose labels GitHub
// refuses is re-filed unlabelled rather than lost (ADR 0017 §8), and
// a deck request whose issue answers 404 or 410 is treated as closed
// (ADR 0095 §3). It satisfies the unexported statusCoder interface
// lobby declares, so the reporter seams stay narrow and a test fake
// can satisfy them without this package.
type APIError struct {
	// Op names the call that failed ("create issue", "get issue",
	// "create comment"). Empty reads as "create issue", the only call
	// this package made before ADR 0095.
	Op     string
	Status int
	Body   string // already truncated
}

func (e *APIError) Error() string {
	op := e.Op
	if op == "" {
		op = "create issue"
	}
	return fmt.Sprintf("github: %s: status %d: %s", op, e.Status, e.Body)
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

	var out struct {
		Number  int    `json:"number"`
		HTMLURL string `json:"html_url"`
	}
	path := fmt.Sprintf("/repos/%s/issues", c.repo)
	if err := c.do(ctx, "create issue", http.MethodPost, path, payload, http.StatusCreated, &out); err != nil {
		return "", 0, err
	}
	return out.HTMLURL, out.Number, nil
}

// Issue is the part of a GitHub issue ADR 0095's deck requests read:
// whether it is still open, and where it lives.
type Issue struct {
	Number  int    `json:"number"`
	State   string `json:"state"` // "open" | "closed"
	HTMLURL string `json:"html_url"`
}

// Open reports whether the issue is open.
func (i Issue) Open() bool { return i.State == "open" }

// GetIssue reads one issue by number. A deleted issue answers 404 or
// 410, surfaced as an *APIError carrying that status so the caller
// can tell "gone" from "GitHub is down".
func (c *Client) GetIssue(ctx context.Context, number int) (Issue, error) {
	var out Issue
	path := fmt.Sprintf("/repos/%s/issues/%d", c.repo, number)
	if err := c.do(ctx, "get issue", http.MethodGet, path, nil, http.StatusOK, &out); err != nil {
		return Issue{}, err
	}
	return out, nil
}

// CreateComment adds a comment to an issue and returns the comment's
// html_url.
func (c *Client) CreateComment(ctx context.Context, number int, body string) (string, error) {
	payload := struct {
		Body string `json:"body"`
	}{Body: body}
	var out struct {
		HTMLURL string `json:"html_url"`
	}
	path := fmt.Sprintf("/repos/%s/issues/%d/comments", c.repo, number)
	if err := c.do(ctx, "create comment", http.MethodPost, path, payload, http.StatusCreated, &out); err != nil {
		return "", err
	}
	return out.HTMLURL, nil
}

// do performs one API call: payload (nil for none) is sent as JSON, a
// status other than want is an *APIError, and the response body —
// capped at maxResponseBytes — is decoded into out. op names the call
// in error strings. The Authorization header never reaches an error:
// http.Client errors carry the URL only.
func (c *Client) do(ctx context.Context, op, method, path string, payload any, want int, out any) error {
	var body io.Reader
	if payload != nil {
		buf, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("github: %s: encode: %w", op, err)
		}
		body = bytes.NewReader(buf)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return fmt.Errorf("github: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.http.Do(req)
	if err != nil {
		// err carries the URL but never the Authorization header,
		// so it is safe to propagate verbatim.
		return fmt.Errorf("github: %s: %w", op, err)
	}
	defer func() { _ = resp.Body.Close() }()

	lim := io.LimitReader(resp.Body, maxResponseBytes)
	if resp.StatusCode != want {
		slurp, _ := io.ReadAll(lim)
		return &APIError{Op: op, Status: resp.StatusCode, Body: truncate(string(slurp), 300)}
	}
	if err := json.NewDecoder(lim).Decode(out); err != nil {
		return fmt.Errorf("github: decode response: %w", err)
	}
	return nil
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
