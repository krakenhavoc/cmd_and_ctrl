package lobby

import (
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
)

// Guard for #721: the in-app bug report published session tokens in
// GitHub issue bodies, because the client log it inlines carried the
// WebSocket URL verbatim. Durable sessions (#517) would keep such a
// token valid across restarts. The client now redacts at the source;
// these tests pin that the server does too, for every client that
// doesn't.

const leakedToken = "Zk3q9_Rb-2xVn8LmPq4tYw7cHs1dJe6uKo0aBf5gTiA"

const leakedInvite = "iNv1te-T0ken_abcdefghijklmnop"

// unredactedToken matches token=<anything but REDACTED>, literal or
// percent-encoded.
var unredactedToken = regexp.MustCompile(`(?i)token(=|%3d)[^R&\s]`)

// assertNoCredential fails if s carries either leaked credential or an
// unredacted token= pair.
func assertNoCredential(t *testing.T, where, s string) {
	t.Helper()
	if strings.Contains(s, leakedToken) {
		t.Errorf("%s contains the session token:\n%s", where, s)
	}
	if strings.Contains(s, leakedInvite) {
		t.Errorf("%s contains the invite token:\n%s", where, s)
	}
	if m := unredactedToken.FindString(s); m != "" {
		t.Errorf("%s contains an unredacted %q:\n%s", where, m, s)
	}
}

// leakyLog is what a pre-#721 client sends: the connect line with the
// token, plus the other shapes a credential turns up in.
func leakyLog() []bugLogEntry {
	wsURL := "wss://cmd.example/ws?game=11111111-2222-3333-4444-555555555555&token=" + leakedToken + "&player=p-1"
	return []bugLogEntry{
		{At: 1_700_000_000_000, Kind: "info", Text: "connected to " + wsURL},
		{At: 1_700_000_000_001, Kind: "console", Text: "console.error: GET /avatars/1/a.png?token=" + leakedToken + " 401"},
		{At: 1_700_000_000_002, Kind: "console", Text: `console.warn: {"token":"` + leakedToken + `"}`},
		{At: 1_700_000_000_003, Kind: "error", Text: "Authorization: Bearer " + leakedToken},
		{At: 1_700_000_000_004, Kind: "info", Text: "open https://h/#/games/abc/join?t=" + leakedInvite},
		{At: 1_700_000_000_005, Kind: "info", Text: "next=" + url.QueryEscape("/ws?token="+leakedToken+"&game=g")},
	}
}

func TestRenderBugIssueBodyNeverContainsAToken(t *testing.T) {
	in := bugIssue{
		Principal: auth.Principal{Role: auth.RolePlayer, Name: "Alice"},
		Kind:      bugKinds["bug"],
		Desc:      "my invite https://h/#/games/abc/join?t=" + leakedInvite + " stopped working (token=" + leakedToken + ")",
		Ctx: &bugReportContext{
			GameID:     uuid.New().String(),
			Turn:       2,
			Phase:      "main1",
			Step:       "main",
			Connection: "wss://h/ws?token=" + leakedToken,
		},
		Log:       leakyLog(),
		UserAgent: "test-browser/1.0 token=" + leakedToken,
		Now:       time.Unix(0, 0).UTC(),
	}
	body := renderBugIssueBody(in)
	assertNoCredential(t, "issue body", body)

	// Redaction must keep what triage needs: the connect line, its game
	// and seat, and the fact that a token was sent.
	for _, want := range []string{
		"connected to wss://cmd.example/ws?game=11111111-2222-3333-4444-555555555555&token=REDACTED&player=p-1",
		"join?t=REDACTED",
		"Bearer REDACTED",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("issue body missing %q\n----\n%s", want, body)
		}
	}

	// The caller's struct is not mutated: redaction works on a copy.
	if !strings.Contains(in.Log[0].Text, leakedToken) || !strings.Contains(in.Ctx.Connection, leakedToken) {
		t.Error("renderBugIssueBody mutated its input")
	}
}

// A log line longer than the clip must not have its key clipped off
// while its value survives: redaction runs before clipping.
func TestRenderBugIssueBodyRedactsBeforeClipping(t *testing.T) {
	pad := strings.Repeat("x", bugLogLineMax-len("connected to /ws?token=")-4)
	in := bugIssue{
		Principal: auth.Principal{Role: auth.RolePlayer, Name: "Alice"},
		Kind:      bugKinds["bug"],
		Log:       []bugLogEntry{{At: 1_700_000_000_000, Kind: "info", Text: pad + "connected to /ws?token=" + leakedToken}},
		Now:       time.Unix(0, 0).UTC(),
	}
	body := renderBugIssueBody(in)
	if strings.Contains(body, leakedToken[:4]) {
		t.Errorf("clipped log line kept a token prefix:\n%s", body)
	}
}

// End to end through the handler: an old client posts a leaky title,
// description and log; nothing that reaches GitHub carries the token.
func TestBugReportHandlerRedactsTokensBeforeFiling(t *testing.T) {
	rep := &recordingReporter{}
	srv, tok := newBugReportStack(t, rep)

	resp := postBugReport(t, srv, tok, map[string]any{
		"title":       "join link ?t=" + leakedInvite + " broken",
		"description": "token=" + leakedToken,
		"context": map[string]any{
			"game_id":    uuid.New().String(),
			"connection": "wss://h/ws?token=" + leakedToken,
		},
		"log": leakyLog(),
	})
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	if len(rep.titles) != 1 {
		t.Fatalf("reporter calls = %d", len(rep.titles))
	}
	assertNoCredential(t, "issue title", rep.titles[0])
	assertNoCredential(t, "issue body", rep.bodies[0])
	if !strings.Contains(rep.bodies[0], "token=REDACTED") {
		t.Errorf("issue body should show a token was redacted:\n%s", rep.bodies[0])
	}
}
