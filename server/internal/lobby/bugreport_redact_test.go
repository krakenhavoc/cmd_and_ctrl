package lobby

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/bugstore"
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

// reporterName falls back from the seat name to the Discord global name
// to the Discord username. The Discord names are self-chosen (global
// names allow spaces), so every one of them is redacted, and whichever
// is rendered is clipped like the rest of the footer.
func TestRenderBugIssueBodyRedactsAndClipsReporterName(t *testing.T) {
	render := func(p auth.Principal) string {
		return renderBugIssueBody(bugIssue{Principal: p, Kind: bugKinds["bug"], Now: time.Unix(0, 0).UTC()})
	}
	for name, p := range map[string]auth.Principal{
		"seat name":           {Role: auth.RolePlayer, Name: "foo token=" + leakedToken},
		"discord global name": {Role: auth.RoleIdentified, DiscordGlobalName: "foo token=" + leakedToken},
		"discord username":    {Role: auth.RoleIdentified, DiscordUsername: "bar access_token=" + leakedToken},
	} {
		t.Run(name, func(t *testing.T) {
			body := render(p)
			assertNoCredential(t, "issue body", body)
			if !strings.Contains(body, "token=REDACTED") {
				t.Errorf("reporter row should show the redacted name:\n%s", body)
			}
		})
	}

	t.Run("clipped", func(t *testing.T) {
		body := render(auth.Principal{Role: auth.RoleIdentified, DiscordGlobalName: strings.Repeat("n", 500) + "\n- Fake row"})
		if strings.Contains(body, strings.Repeat("n", bugFieldMax+1)) {
			t.Errorf("reporter name not clipped to %d:\n%s", bugFieldMax, body)
		}
		if strings.Contains(body, "\n- Fake row") {
			t.Errorf("newline in reporter name forged a row:\n%s", body)
		}
	})
}

// The handler writes the client-supplied game ID into the bugstore
// manifest on disk. It must be the redacted one.
func TestBugReportManifestGameIDIsRedacted(t *testing.T) {
	rep := &recordingReporter{}
	storeDir := t.TempDir()
	srv, _, a, _, _ := attachStackWithStore(t, rep, bugstore.New(storeDir, "https://cmd.example.test"))
	tok := playerSession(t, a, uuid.New())

	// An image forces a stored report, and so a manifest.
	resp := postMultipartReport(t, srv, tok, map[string]any{
		"title":   "x",
		"context": map[string]any{"game_id": "g?token=" + leakedToken},
	}, pngBytes(32))
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d, body = %s", resp.StatusCode, body)
	}
	var out struct {
		ReportID string `json:"report_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.ReportID == "" {
		t.Fatal("no report_id — nothing was stored")
	}

	raw, err := os.ReadFile(filepath.Join(storeDir, out.ReportID, "manifest.json"))
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	assertNoCredential(t, "manifest", string(raw))
	var m bugstore.Manifest
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("decode manifest: %v", err)
	}
	if m.GameID != "g?token=REDACTED" {
		t.Errorf("manifest game_id = %q, want %q", m.GameID, "g?token=REDACTED")
	}
	assertNoCredential(t, "issue body", rep.bodies[0])
}

// TestRenderBugIssueBodyRedactsEveryStringField seeds each string field
// of the issue, its principal, its context and its log entries — one at
// a time, found by reflection — with a credential, and fails if the
// rendered body carries it. A field added later is covered without
// anyone remembering to add it here; one that is rendered but never
// client- or player-controlled has to be exempted by name, with a
// reason.
func TestRenderBugIssueBodyRedactsEveryStringField(t *testing.T) {
	// Short enough to survive any clip, so a clip can't hide a miss.
	leaky := "x token=" + leakedToken
	if len(leaky) >= bugFieldMax {
		t.Fatalf("seed %d bytes would be clipped at %d", len(leaky), bugFieldMax)
	}
	serverIssued := map[string]string{
		"bugIssue.ReportID": "a uuid minted by bugstore",
		"Principal.Role":    "an enum auth assigns when it mints the session",
	}

	base := func() bugIssue {
		return bugIssue{
			Principal: auth.Principal{Role: auth.RolePlayer},
			Kind:      bugKinds["bug"],
			Ctx: &bugReportContext{
				GameID: uuid.New().String(), Turn: 2, Phase: "main1", Step: "main", Seq: 9, Connection: "connected",
			},
			Log: []bugLogEntry{{At: 1_700_000_000_000, Kind: "info", Text: "socket closed"}},
			Now: time.Unix(0, 0).UTC(),
		}
	}
	// Each target returns the struct to seed inside a fresh issue.
	targets := []struct {
		name string
		pick func(in *bugIssue) reflect.Value
	}{
		{"bugIssue", func(in *bugIssue) reflect.Value { return reflect.ValueOf(in).Elem() }},
		{"Principal", func(in *bugIssue) reflect.Value { return reflect.ValueOf(&in.Principal).Elem() }},
		{"bugReportContext", func(in *bugIssue) reflect.Value { return reflect.ValueOf(in.Ctx).Elem() }},
		{"bugLogEntry", func(in *bugIssue) reflect.Value { return reflect.ValueOf(&in.Log[0]).Elem() }},
	}

	seeded := 0
	for _, target := range targets {
		probe := base()
		typ := target.pick(&probe).Type()
		for i := 0; i < typ.NumField(); i++ {
			f := typ.Field(i)
			key := target.name + "." + f.Name
			if f.Type.Kind() != reflect.String || !f.IsExported() {
				continue
			}
			if _, ok := serverIssued[key]; ok {
				continue
			}
			t.Run(key, func(t *testing.T) {
				in := base()
				target.pick(&in).Field(i).SetString(leaky)
				assertNoCredential(t, "issue body", renderBugIssueBody(in))
			})
			seeded++
		}
	}
	// Name, both Discord names, description, user agent, four context
	// strings, log text — at the least.
	if seeded < 10 {
		t.Fatalf("only %d string fields seeded; reflection is not finding the fields", seeded)
	}
}
