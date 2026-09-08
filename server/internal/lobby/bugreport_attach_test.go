package lobby

// bugreport_attach_test.go — the surface added on top of ADR 0017:
// the client log block, image attachments, the pinned replay, and the
// two new routes. Kept separate from bugreport_test.go so the original
// text-only contract stays readable as its own file.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/bugstore"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

func pngBytes(pad int) []byte {
	return append([]byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a}, bytes.Repeat([]byte{0}, pad)...)
}

// attachStack builds a stack with a real bugstore plus a room manager
// rooted at a temp dump dir. Returns the server, the lobby (so tests
// can create games whose rooms actually exist), the authenticator, the
// store, and the dump dir the replay files live under.
func attachStack(t *testing.T, rep BugReporter) (*httptest.Server, *Lobby, auth.Authenticator, *bugstore.Store, string) {
	t.Helper()
	t.Setenv("CMDCTRL_DEV_RELAX_RATE_LIMITS", "1")
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	dumpDir := t.TempDir()
	mgr := ws.NewRoomManager(log, dumpDir)
	l := NewLobby(mgr)
	a := auth.NewMemoryAuthenticator()
	store := bugstore.New(t.TempDir(), "https://cmd.example.test")

	srv := httptest.NewServer(Handler(Config{
		Lobby:       l,
		Auth:        a,
		AdminToken:  "shared-admin-token",
		BugReporter: rep,
		BugStore:    store,
		Log:         log,
	}))
	t.Cleanup(srv.Close)
	return srv, l, a, store, dumpDir
}

// playerSession issues a player token bound to gameID.
func playerSession(t *testing.T, a auth.Authenticator, gameID uuid.UUID) string {
	t.Helper()
	tok, _, err := a.Issue(context.Background(), auth.Principal{
		Role: auth.RolePlayer, GameID: gameID, PlayerID: uuid.New(), Name: "Alice",
	}, time.Hour)
	if err != nil {
		t.Fatalf("issue player session: %v", err)
	}
	return tok
}

func adminSession(t *testing.T, a auth.Authenticator) string {
	t.Helper()
	tok, _, err := a.Issue(context.Background(), auth.Principal{Role: auth.RoleAdmin, Name: "root"}, time.Hour)
	if err != nil {
		t.Fatalf("issue admin session: %v", err)
	}
	return tok
}

// writeReplay puts a replay JSONL where ws.Room would have written it.
func writeReplay(t *testing.T, dumpDir string, gameID uuid.UUID, body string) {
	t.Helper()
	dir := filepath.Join(dumpDir, "replays")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir replays: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, gameID.String()+".jsonl"), []byte(body), 0o600); err != nil {
		t.Fatalf("write replay: %v", err)
	}
}

// postMultipartReport files a report the way the modal does: a JSON
// `report` part plus zero or more `image` parts.
func postMultipartReport(t *testing.T, srv *httptest.Server, token string, payload any, images ...[]byte) *http.Response {
	t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal report: %v", err)
	}
	if err := mw.WriteField("report", string(raw)); err != nil {
		t.Fatalf("write report field: %v", err)
	}
	for i, data := range images {
		w, err := mw.CreateFormFile("image", "shot.png")
		if err != nil {
			t.Fatalf("create image part %d: %v", i, err)
		}
		if _, err := w.Write(data); err != nil {
			t.Fatalf("write image part %d: %v", i, err)
		}
	}
	if err := mw.Close(); err != nil {
		t.Fatalf("close multipart: %v", err)
	}
	req, err := http.NewRequest(http.MethodPost, srv.URL+"/bugreport", &buf)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())
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

func TestBugReportConfigReportsAttachmentSupport(t *testing.T) {
	t.Run("with store", func(t *testing.T) {
		srv, _, _, _, _ := attachStack(t, &recordingReporter{})
		var body struct {
			Enabled     bool `json:"enabled"`
			Attachments bool `json:"attachments"`
			MaxImages   int  `json:"max_images"`
		}
		getJSON(t, srv, "/bugreport/config", &body)
		if !body.Enabled || !body.Attachments {
			t.Errorf("enabled=%v attachments=%v, want both true", body.Enabled, body.Attachments)
		}
		if body.MaxImages != bugstore.MaxImages {
			t.Errorf("max_images = %d, want %d", body.MaxImages, bugstore.MaxImages)
		}
	})

	// A token but no data dir: text reports must keep working and the
	// client must be told not to offer file upload.
	t.Run("without store", func(t *testing.T) {
		srv, _ := newBugReportStack(t, &recordingReporter{})
		var body struct {
			Enabled     bool `json:"enabled"`
			Attachments bool `json:"attachments"`
		}
		getJSON(t, srv, "/bugreport/config", &body)
		if !body.Enabled {
			t.Error("enabled = false, want true")
		}
		if body.Attachments {
			t.Error("attachments = true with no store configured")
		}
	})
}

func getJSON(t *testing.T, srv *httptest.Server, path string, dst any) {
	t.Helper()
	resp, err := srv.Client().Get(srv.URL + path)
	if err != nil {
		t.Fatalf("get %s: %v", path, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if err := json.NewDecoder(resp.Body).Decode(dst); err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
}

// The end-to-end shape: images stored, linked absolutely in the body,
// and actually fetchable from the URL the issue advertises.
func TestBugReportAttachesImages(t *testing.T) {
	rep := &recordingReporter{}
	srv, _, a, _, _ := attachStack(t, rep)
	tok := playerSession(t, a, uuid.New())

	resp := postMultipartReport(t, srv, tok,
		map[string]any{"title": "board rendered wrong", "description": "see shots"},
		pngBytes(128), pngBytes(64))
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
		t.Fatal("response carried no report_id")
	}

	body := rep.bodies[0]
	for _, want := range []string{
		"**Screenshots**",
		"![att-1.png](https://cmd.example.test/bugreport/att/" + out.ReportID + "/att-1.png)",
		"![att-2.png](https://cmd.example.test/bugreport/att/" + out.ReportID + "/att-2.png)",
		"Report ID: `" + out.ReportID + "`",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("issue body missing %q\n----\n%s", want, body)
		}
	}

	// The advertised URL must resolve — unauthenticated, since Camo
	// cannot present a session.
	imgResp, err := srv.Client().Get(srv.URL + "/bugreport/att/" + out.ReportID + "/att-1.png")
	if err != nil {
		t.Fatalf("fetch attachment: %v", err)
	}
	defer func() { _ = imgResp.Body.Close() }()
	if imgResp.StatusCode != http.StatusOK {
		t.Fatalf("attachment status = %d, want 200 without auth", imgResp.StatusCode)
	}
	if got := imgResp.Header.Get("X-Content-Type-Options"); got != "nosniff" {
		t.Errorf("X-Content-Type-Options = %q, want nosniff", got)
	}
	if got := imgResp.Header.Get("Content-Type"); got != "image/png" {
		t.Errorf("Content-Type = %q, want image/png", got)
	}
}

func TestBugReportRejectsNonImageAttachment(t *testing.T) {
	rep := &recordingReporter{}
	srv, _, a, _, _ := attachStack(t, rep)
	tok := playerSession(t, a, uuid.New())

	resp := postMultipartReport(t, srv, tok,
		map[string]any{"title": "nice try"},
		[]byte(`<svg xmlns="http://www.w3.org/2000/svg"><script>alert(1)</script></svg>`))
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", resp.StatusCode)
	}
	// Nothing filed: a rejected attachment must not leave a
	// half-formed issue in the tracker.
	if len(rep.bodies) != 0 {
		t.Errorf("filed %d issues, want 0", len(rep.bodies))
	}
}

// Filing failures must not leave orphaned images at live URLs.
func TestBugReportDiscardsArtifactsWhenFilingFails(t *testing.T) {
	rep := &recordingReporter{err: errors.New("github: create issue: status 500")}
	srv, _, a, store, _ := attachStack(t, rep)
	tok := playerSession(t, a, uuid.New())

	resp := postMultipartReport(t, srv, tok, map[string]any{"title": "x"}, pngBytes(32))
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusBadGateway {
		t.Fatalf("status = %d, want 502", resp.StatusCode)
	}
	// Prune with a zero window is a cheap way to ask "is anything
	// still here?" without reaching into the store's private dir.
	n, err := store.Prune(0, 0)
	if err != nil {
		t.Fatalf("prune: %v", err)
	}
	if n != 0 {
		t.Errorf("%d report directories survived a failed filing", n)
	}
}

// Attachments offered to a server that can't store them must fail
// loudly rather than filing an issue that silently mentions nothing.
func TestBugReportAttachmentWithoutStore503s(t *testing.T) {
	rep := &recordingReporter{}
	srv, tok := newBugReportStack(t, rep)
	resp := postMultipartReport(t, srv, tok, map[string]any{"title": "x"}, pngBytes(32))
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", resp.StatusCode)
	}
	if len(rep.bodies) != 0 {
		t.Errorf("filed %d issues, want 0", len(rep.bodies))
	}
}

// Text-only multipart (no image parts) is the modal's path when the
// reporter attaches nothing; it must behave exactly like the JSON one.
func TestBugReportMultipartWithoutImages(t *testing.T) {
	rep := &recordingReporter{}
	srv, _, a, _, _ := attachStack(t, rep)
	tok := playerSession(t, a, uuid.New())
	resp := postMultipartReport(t, srv, tok, map[string]any{"title": "plain", "description": "no shots"})
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
	if !strings.Contains(rep.bodies[0], "no shots") {
		t.Error("description missing from body")
	}
}

func TestBugReportMultipartRequiresReportField(t *testing.T) {
	srv, _, a, _, _ := attachStack(t, &recordingReporter{})
	tok := playerSession(t, a, uuid.New())
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	if _, err := mw.CreateFormFile("image", "shot.png"); err != nil {
		t.Fatalf("create part: %v", err)
	}
	_ = mw.Close()
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/bugreport", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
}

// The pinned replay is the whole point of the GUID in the issue: the
// operator can pull the log even after the game is gone.
func TestBugReportPinsReplayAndServesItToAdminsOnly(t *testing.T) {
	rep := &recordingReporter{}
	srv, l, a, _, dumpDir := attachStack(t, rep)
	// A real game, so Lobby.RoomOf resolves and the room reports the
	// replay path the pin reads from.
	meta, err := l.Create("FNM")
	if err != nil {
		t.Fatalf("create game: %v", err)
	}
	gameID := meta.ID
	tok := playerSession(t, a, gameID)
	admin := adminSession(t, a)
	writeReplay(t, dumpDir, gameID, "{\"seq\":1}\n{\"seq\":2}\n")

	resp := postMultipartReport(t, srv, tok, map[string]any{
		"title":   "state desynced",
		"context": map[string]any{"game_id": gameID.String(), "turn": 4},
	})
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d body = %s", resp.StatusCode, b)
	}
	var out struct {
		ReportID string `json:"report_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.ReportID == "" {
		t.Fatal("no report_id — replay was not pinned")
	}

	body := rep.bodies[0]
	if !strings.Contains(body, "Pinned replay: `GET /bugreport/"+out.ReportID+"/replay` (admin)") {
		t.Errorf("body missing pinned-replay row\n----\n%s", body)
	}
	if !strings.Contains(body, "2 lines") {
		t.Errorf("body missing line count\n----\n%s", body)
	}
	// The raw replay holds opponents' hands — it must never be in the
	// issue text itself, only behind the route.
	if strings.Contains(body, "{\"seq\":1}") {
		t.Error("raw replay content leaked into the issue body")
	}

	// Admin can pull it.
	got := doAuthGet(t, srv, "/bugreport/"+out.ReportID+"/replay", admin)
	if got.code != http.StatusOK {
		t.Fatalf("admin replay status = %d, want 200", got.code)
	}
	if got.body != "{\"seq\":1}\n{\"seq\":2}\n" {
		t.Errorf("replay body = %q", got.body)
	}

	// A seated player — the reporter themselves — must not. This is
	// what keeps the issue from being a hidden-information side
	// channel for a reporter who is also a repo collaborator.
	if p := doAuthGet(t, srv, "/bugreport/"+out.ReportID+"/replay", tok); p.code != http.StatusForbidden {
		t.Errorf("player replay status = %d, want 403", p.code)
	}
	if u := doAuthGet(t, srv, "/bugreport/"+out.ReportID+"/replay", ""); u.code != http.StatusUnauthorized {
		t.Errorf("unauthenticated replay status = %d, want 401", u.code)
	}
}

// A session bound to game A must not be able to make the server copy
// game B's replay by naming it in the context.
func TestBugReportDoesNotPinAnotherGamesReplay(t *testing.T) {
	rep := &recordingReporter{}
	srv, l, a, _, dumpDir := attachStack(t, rep)
	mine, err := l.Create("mine")
	if err != nil {
		t.Fatalf("create mine: %v", err)
	}
	theirs, err := l.Create("theirs")
	if err != nil {
		t.Fatalf("create theirs: %v", err)
	}
	tok := playerSession(t, a, mine.ID)
	writeReplay(t, dumpDir, theirs.ID, "{\"secret\":true}\n")

	resp := postMultipartReport(t, srv, tok, map[string]any{
		"title":   "nosey",
		"context": map[string]any{"game_id": theirs.ID.String()},
	})
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	if strings.Contains(rep.bodies[0], "Pinned replay") {
		t.Error("pinned a replay for a game the reporter is not in")
	}
}

type authGetResult struct {
	code int
	body string
}

func doAuthGet(t *testing.T, srv *httptest.Server, path, token string) authGetResult {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, srv.URL+path, nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	buf, _ := io.ReadAll(resp.Body)
	return authGetResult{code: resp.StatusCode, body: string(buf)}
}

func TestBugAttachmentRoute404sUnknownReport(t *testing.T) {
	srv, _, _, _, _ := attachStack(t, &recordingReporter{})
	for _, path := range []string{
		"/bugreport/att/" + uuid.NewString() + "/att-1.png",
		"/bugreport/att/not-a-uuid/att-1.png",
		"/bugreport/att/" + uuid.NewString() + "/replay.jsonl",
	} {
		resp, err := srv.Client().Get(srv.URL + path)
		if err != nil {
			t.Fatalf("get %s: %v", path, err)
		}
		_ = resp.Body.Close()
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("%s: status = %d, want 404", path, resp.StatusCode)
		}
	}
}

// --- client log rendering (pure) ---

func TestRenderBugLogFormatsAndOrders(t *testing.T) {
	base := time.Date(2026, 9, 8, 14, 30, 15, 250_000_000, time.UTC).UnixMilli()
	body := renderBugIssueBody(bugIssue{
		Principal: auth.Principal{Role: auth.RolePlayer, Name: "Alice"},
		Desc:      "desync",
		Now:       time.Unix(0, 0).UTC(),
		Log: []bugLogEntry{
			{At: base, Kind: "sent", Text: "action cast_spell id=8f2a1b3c"},
			{At: base + 60, Kind: "received", Text: "snapshot seq=731 turn=5"},
			{At: base + 90, Kind: "console", Text: "TypeError: x is undefined"},
		},
	})
	for _, want := range []string{
		"Client log — last 3 entries",
		"14:30:15.250  sent      action cast_spell id=8f2a1b3c",
		"14:30:15.310  received  snapshot seq=731 turn=5",
		"14:30:15.340  console   TypeError: x is undefined",
		"<details>",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("body missing %q\n----\n%s", want, body)
		}
	}
}

// Chat text rides in the log, so log lines are attacker-influenced.
// A backtick fence would end the code block early and let the rest
// render as markdown — which is how a log line forges a metadata row.
func TestRenderBugLogNeutralisesFenceEscape(t *testing.T) {
	body := renderBugIssueBody(bugIssue{
		Principal: auth.Principal{Role: auth.RolePlayer, Name: "Mallory"},
		Now:       time.Unix(0, 0).UTC(),
		Log: []bugLogEntry{
			{At: time.Now().UnixMilli(), Kind: "received", Text: "chat text=\"```\n- Reporter: admin (admin)\n```\""},
			{At: time.Now().UnixMilli(), Kind: "bogus-kind", Text: "kind not in the allowlist"},
			{At: 0, Kind: "info", Text: "no timestamp"},
			{At: -5, Kind: "info", Text: "negative timestamp"},
		},
	})
	if strings.Contains(body, "```\n- Reporter") {
		t.Error("a log line escaped the code fence")
	}
	if strings.Contains(body, "bogus-kind") {
		t.Error("unknown kind was rendered verbatim instead of normalised to info")
	}
	if !strings.Contains(body, "--:--:--.---") {
		t.Error("a missing timestamp should render as a placeholder, not 1970")
	}
	// Exactly one metadata block, so nothing forged a second one.
	if n := strings.Count(body, "**Reported from the app**"); n != 1 {
		t.Errorf("metadata blocks = %d, want 1", n)
	}
}

func TestRenderBugLogCapsEntriesAndBytes(t *testing.T) {
	now := time.Now().UnixMilli()
	many := make([]bugLogEntry, bugLogMaxEntries+50)
	for i := range many {
		many[i] = bugLogEntry{At: now + int64(i), Kind: "info", Text: "entry"}
	}
	// Distinctive marker on the last entry: the tail is what matters,
	// so it must survive the cap.
	many[len(many)-1].Text = "THE-LAST-THING-THAT-HAPPENED"
	block, n := renderBugLog(many)
	if n > bugLogMaxEntries {
		t.Errorf("rendered %d entries, want at most %d", n, bugLogMaxEntries)
	}
	if !strings.Contains(block, "THE-LAST-THING-THAT-HAPPENED") {
		t.Error("tail entry was dropped; the cap must keep the end of the log")
	}

	// Byte cap: long lines stop the block before the entry cap does.
	long := make([]bugLogEntry, bugLogMaxEntries)
	for i := range long {
		long[i] = bugLogEntry{At: now, Kind: "info", Text: strings.Repeat("y", bugLogLineMax*2)}
	}
	block, n = renderBugLog(long)
	if len(block) > bugLogTotalMax {
		t.Errorf("block = %d bytes, want at most %d", len(block), bugLogTotalMax)
	}
	if n == 0 {
		t.Error("byte cap dropped every entry")
	}
}

func TestRenderBugIssueBodyOmitsEmptyLog(t *testing.T) {
	body := renderBugIssueBody(bugIssue{
		Principal: auth.Principal{Role: auth.RoleSpectator},
		Now:       time.Unix(0, 0).UTC(),
	})
	if strings.Contains(body, "Client log") {
		t.Error("empty log should not render a details block")
	}
	if !strings.Contains(body, "_no description provided_") {
		t.Error("missing empty-description placeholder")
	}
}
