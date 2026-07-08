package lobby

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// recordingReporter is the BugReporter fake: records every call and
// returns a canned issue (or error).
type recordingReporter struct {
	titles []string
	bodies []string
	labels [][]string
	err    error
}

func (r *recordingReporter) CreateIssue(_ context.Context, title, body string, labels []string) (string, int, error) {
	if r.err != nil {
		return "", 0, r.err
	}
	r.titles = append(r.titles, title)
	r.bodies = append(r.bodies, body)
	r.labels = append(r.labels, labels)
	return "https://github.com/o/r/issues/7", 7, nil
}

// newBugReportStack builds a minimal HTTP stack with the reporter
// wired (nil = disabled) and returns the server plus a player
// session token for gameID-less reports.
func newBugReportStack(t *testing.T, rep BugReporter) (*httptest.Server, string) {
	t.Helper()
	// The bug limiter's burst (3) is smaller than some tests' call
	// counts; relax limits the same way the e2e suite does. Read at
	// Handler-construction time, so set before building the stack.
	t.Setenv("CMDCTRL_DEV_RELAX_RATE_LIMITS", "1")
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	mgr := ws.NewRoomManager(log, "")
	l := NewLobby(mgr)
	a := auth.NewMemoryAuthenticator()

	srv := httptest.NewServer(Handler(Config{
		Lobby:       l,
		Auth:        a,
		AdminToken:  "shared-admin-token",
		BugReporter: rep,
	}))
	t.Cleanup(srv.Close)

	tok, _, err := a.Issue(context.Background(), auth.Principal{
		Role:     auth.RolePlayer,
		GameID:   uuid.New(),
		PlayerID: uuid.New(),
		Name:     "Alice",
	}, time.Hour)
	if err != nil {
		t.Fatalf("issue session: %v", err)
	}
	return srv, tok
}

func postBugReport(t *testing.T, srv *httptest.Server, token string, payload any) *http.Response {
	t.Helper()
	buf, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	req, err := http.NewRequest(http.MethodPost, srv.URL+"/bugreport", bytes.NewReader(buf))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "test-browser/1.0")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	return resp
}

func TestBugReportConfigFlag(t *testing.T) {
	for _, tc := range []struct {
		name string
		rep  BugReporter
		want bool
	}{
		{"disabled", nil, false},
		{"enabled", &recordingReporter{}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv, _ := newBugReportStack(t, tc.rep)
			resp, err := srv.Client().Get(srv.URL + "/bugreport/config")
			if err != nil {
				t.Fatalf("get: %v", err)
			}
			defer func() { _ = resp.Body.Close() }()
			var body struct {
				Enabled bool `json:"enabled"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if body.Enabled != tc.want {
				t.Fatalf("enabled = %v, want %v", body.Enabled, tc.want)
			}
		})
	}
}

func TestBugReportFilesIssue(t *testing.T) {
	rep := &recordingReporter{}
	srv, tok := newBugReportStack(t, rep)

	gameID := uuid.New().String()
	resp := postBugReport(t, srv, tok, map[string]any{
		"title":       "  cast dialog eats X value  ",
		"description": "set X=4, dialog sent X=0",
		"context": map[string]any{
			"game_id":    gameID,
			"turn":       5,
			"phase":      "main1",
			"step":       "main",
			"seq":        731,
			"connection": "connected",
		},
	})
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	var out struct {
		URL    string `json:"url"`
		Number int    `json:"number"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.Number != 7 || !strings.Contains(out.URL, "/issues/7") {
		t.Fatalf("unexpected response: %+v", out)
	}

	if len(rep.titles) != 1 {
		t.Fatalf("reporter calls = %d", len(rep.titles))
	}
	if rep.titles[0] != "[in-app] cast dialog eats X value" {
		t.Errorf("title = %q (want prefix + trimmed)", rep.titles[0])
	}
	body := rep.bodies[0]
	for _, want := range []string{
		"set X=4, dialog sent X=0",
		"Reporter: Alice (player)",
		"`" + gameID + "`",
		"turn 5, main1/main",
		"seq 731",
		"GET /games/" + gameID + "/replay",
		"User agent: test-browser/1.0",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("issue body missing %q\n----\n%s", want, body)
		}
	}
	if len(rep.labels[0]) != 1 || rep.labels[0][0] != "bug" {
		t.Errorf("labels = %v", rep.labels[0])
	}
}

func TestBugReportRequiresSession(t *testing.T) {
	srv, _ := newBugReportStack(t, &recordingReporter{})
	resp := postBugReport(t, srv, "", map[string]any{"title": "x"})
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
}

func TestBugReportValidation(t *testing.T) {
	rep := &recordingReporter{}
	srv, tok := newBugReportStack(t, rep)

	for name, payload := range map[string]map[string]any{
		"empty title":    {"title": "   "},
		"title too long": {"title": strings.Repeat("t", bugTitleMax+1)},
		"desc too long":  {"title": "ok", "description": strings.Repeat("d", bugDescMax+1)},
		"unknown field":  {"title": "ok", "surprise": true},
	} {
		t.Run(name, func(t *testing.T) {
			resp := postBugReport(t, srv, tok, payload)
			defer func() { _ = resp.Body.Close() }()
			if resp.StatusCode != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400", resp.StatusCode)
			}
		})
	}
	if len(rep.titles) != 0 {
		t.Fatalf("no report should have reached the reporter, got %d", len(rep.titles))
	}
}

func TestBugReportDisabled503(t *testing.T) {
	srv, tok := newBugReportStack(t, nil)
	resp := postBugReport(t, srv, tok, map[string]any{"title": "x"})
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", resp.StatusCode)
	}
}

func TestBugReportUpstreamFailureIs502(t *testing.T) {
	srv, tok := newBugReportStack(t, &recordingReporter{err: errors.New("github: create issue: status 500")})
	resp := postBugReport(t, srv, tok, map[string]any{"title": "x"})
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusBadGateway {
		t.Fatalf("status = %d, want 502", resp.StatusCode)
	}
}

func TestRenderBugIssueBodyClipsHostileContext(t *testing.T) {
	p := auth.Principal{Role: auth.RolePlayer, Name: "Mallory"}
	bctx := &bugReportContext{
		GameID: strings.Repeat("a", 200) + "\n- Fake row",
		Turn:   3,
		Phase:  "main1\ninjected",
		Step:   "main",
	}
	body := renderBugIssueBody(p, "desc", bctx, "", time.Unix(0, 0).UTC())
	if strings.Contains(body, "\n- Fake row") {
		t.Error("newline in game_id must not create a fake list row")
	}
	if strings.Contains(body, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa") {
		t.Error("game_id must be clipped to bugFieldMax")
	}
}
