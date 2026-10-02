package lobby

// Tests for ADR 0110 section 4's account settings routes, GET and PUT
// /me/settings (Delivery PR 5): who may call them, the revision check,
// the old-client refusal, the size cap and the rate limit.

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/usersettings"
)

// settingsReply is the body of every /me/settings answer that carries
// a copy: GET, a successful PUT, and the 412 and 409 refusals.
type settingsReply struct {
	Error     string          `json:"error"`
	Version   int             `json:"version"`
	Revision  int64           `json:"revision"`
	Settings  json.RawMessage `json:"settings"`
	UpdatedAt int64           `json:"updated_at"`
}

func getSettings(t *testing.T, s userStack, tok string) (int, settingsReply) {
	t.Helper()
	resp := doGet(t, s.srv, "/me/settings", tok)
	return readSettingsReply(t, resp)
}

// putSettings sends PUT /me/settings. ifMatch "" sends no If-Match.
func putSettings(t *testing.T, s userStack, tok, ifMatch string, body string) (int, settingsReply, http.Header) {
	t.Helper()
	req, err := http.NewRequest(http.MethodPut, s.srv.URL+"/me/settings", strings.NewReader(body))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	}
	if ifMatch != "" {
		req.Header.Set("If-Match", ifMatch)
	}
	resp, err := s.srv.Client().Do(req)
	if err != nil {
		t.Fatalf("PUT /me/settings: %v", err)
	}
	status, reply := readSettingsReply(t, resp)
	return status, reply, resp.Header
}

func readSettingsReply(t *testing.T, resp *http.Response) (int, settingsReply) {
	t.Helper()
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	var out settingsReply
	if len(bytes.TrimSpace(raw)) > 0 {
		if err := json.Unmarshal(raw, &out); err != nil {
			t.Fatalf("decode /me/settings (%d): %v (%s)", resp.StatusCode, err, raw)
		}
	}
	return resp.StatusCode, out
}

func settingsBody(version int, settings string) string {
	return `{"version":` + strconv.Itoa(version) + `,"settings":` + settings + `}`
}

// signedInSettingsUser walks a real Discord sign-in and returns its
// session and user.
func signedInSettingsUser(t *testing.T, s userStack) (string, uuid.UUID) {
	t.Helper()
	tok := identityTokenFromCallback(t, s.srv, s.state)
	return tok, mustValidate(t, s.auth, tok).UserID
}

func TestMeSettingsWithNoCopyIsRevisionZero(t *testing.T) {
	s := newMyGamesStack(t)
	tok, _ := signedInSettingsUser(t, s)

	status, got := getSettings(t, s, tok)
	if status != http.StatusOK {
		t.Fatalf("status %d", status)
	}
	if got.Revision != 0 || got.Settings != nil || got.Version != 0 {
		t.Errorf("GET with no row = %+v, want {revision: 0} only", got)
	}
}

func TestMeSettingsRoundTrip(t *testing.T) {
	s := newMyGamesStack(t)
	tok, _ := signedInSettingsUser(t, s)

	status, put, _ := putSettings(t, s, tok, "0", settingsBody(15, `{"display":{"theme":"light"}}`))
	if status != http.StatusOK {
		t.Fatalf("first PUT: %d %+v", status, put)
	}
	if put.Revision != 1 || put.Version != 15 || string(put.Settings) != `{"display":{"theme":"light"}}` || put.UpdatedAt <= 0 {
		t.Errorf("first PUT = %+v", put)
	}

	// A quoted ETag form is accepted too.
	status, put, _ = putSettings(t, s, tok, `"1"`, settingsBody(15, `{"display":{"theme":"dark"}}`))
	if status != http.StatusOK || put.Revision != 2 {
		t.Fatalf("second PUT: %d %+v", status, put)
	}

	status, got := getSettings(t, s, tok)
	if status != http.StatusOK || got.Revision != 2 || got.Version != 15 || string(got.Settings) != `{"display":{"theme":"dark"}}` {
		t.Errorf("GET = %d %+v", status, got)
	}
}

// TestMeSettingsAreTheCallersOwn: a seat session for the same person
// sees the same copy, and another person sees none of it.
func TestMeSettingsAreTheCallersOwn(t *testing.T) {
	s := newMyGamesStack(t)
	tok, me := signedInSettingsUser(t, s)
	if status, _, _ := putSettings(t, s, tok, "0", settingsBody(15, `{"a":1}`)); status != http.StatusOK {
		t.Fatalf("PUT: %d", status)
	}

	seat := playerTokenWithUser(t, s.auth, uuid.New(), uuid.New(), me, "Alice")
	if status, got := getSettings(t, s, seat); status != http.StatusOK || got.Revision != 1 || string(got.Settings) != `{"a":1}` {
		t.Errorf("same person's seat session: %d %+v", status, got)
	}

	other := identityToken(t, s.auth, dmTarget(t, s, "d-other", "Other").ID, "d-other", "Other")
	if status, got := getSettings(t, s, other); status != http.StatusOK || got.Revision != 0 || got.Settings != nil {
		t.Errorf("another person: %d %+v, want no copy", status, got)
	}
}

func TestMeSettingsRefusesEveryCallerWhoIsNotAPerson(t *testing.T) {
	s := newMyGamesStack(t)

	meta, err := s.lobby.Create("guests only")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	_, playerID, err := s.lobby.Join(meta.ID, meta.InviteToken, "Guest")
	if err != nil {
		t.Fatalf("Join: %v", err)
	}
	spectator, _, err := s.auth.Issue(context.Background(), auth.Principal{
		Role: auth.RoleSpectator, GameID: meta.ID, Name: "Watcher",
	}, time.Hour)
	if err != nil {
		t.Fatalf("issue spectator: %v", err)
	}
	callers := []struct {
		name string
		tok  string
		want int
	}{
		{"no session", "", http.StatusUnauthorized},
		{"admin token", adminToken(t, s.srv), http.StatusForbidden},
		{"guest seat", playerToken(t, s.auth, meta.ID, playerID, "Guest"), http.StatusForbidden},
		{"guest spectator", spectator, http.StatusForbidden},
	}
	for _, c := range callers {
		if status, _ := getSettings(t, s, c.tok); status != c.want {
			t.Errorf("GET as %s: %d, want %d", c.name, status, c.want)
		}
		if status, _, _ := putSettings(t, s, c.tok, "0", settingsBody(15, `{}`)); status != c.want {
			t.Errorf("PUT as %s: %d, want %d", c.name, status, c.want)
		}
	}
	// Nothing was written for anybody.
	var n int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM user_settings`).Scan(&n); err != nil || n != 0 {
		t.Errorf("user_settings rows = %d (%v), want 0", n, err)
	}
}

// TestMeSettingsOnADeploymentWithNoDatabase: a Discord sign-in there
// carries no user, so both routes answer 403 and the client stays
// browser-only.
func TestMeSettingsOnADeploymentWithNoDatabase(t *testing.T) {
	srv, _, _, state := newDiscordTestStack(t)
	tok := identityTokenFromCallback(t, srv, state)
	resp := doGet(t, srv, "/me/settings", tok)
	resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("GET: %d, want 403", resp.StatusCode)
	}
}

func TestMeSettingsRevisionConflictReturnsTheCurrentCopy(t *testing.T) {
	s := newMyGamesStack(t)
	tok, _ := signedInSettingsUser(t, s)
	if status, _, _ := putSettings(t, s, tok, "0", settingsBody(15, `{"from":"laptop"}`)); status != http.StatusOK {
		t.Fatal("first PUT")
	}
	if status, _, _ := putSettings(t, s, tok, "1", settingsBody(15, `{"from":"phone"}`)); status != http.StatusOK {
		t.Fatal("second PUT")
	}

	// The laptop still thinks the revision is 1.
	status, got, _ := putSettings(t, s, tok, "1", settingsBody(15, `{"from":"laptop again"}`))
	if status != http.StatusPreconditionFailed {
		t.Fatalf("stale If-Match: %d, want 412", status)
	}
	if got.Revision != 2 || string(got.Settings) != `{"from":"phone"}` || got.Version != 15 || got.Error == "" {
		t.Errorf("412 body = %+v, want the current copy and an error", got)
	}
	// A "first copy" write over an existing row is a conflict too.
	if status, got, _ := putSettings(t, s, tok, "0", settingsBody(15, `{}`)); status != http.StatusPreconditionFailed || got.Revision != 2 {
		t.Errorf("If-Match 0 over a row: %d %+v", status, got)
	}
	if _, cur := getSettings(t, s, tok); string(cur.Settings) != `{"from":"phone"}` {
		t.Errorf("a refused write changed the copy: %s", cur.Settings)
	}
}

func TestMeSettingsRefusesAnOlderClientVersion(t *testing.T) {
	s := newMyGamesStack(t)
	tok, _ := signedInSettingsUser(t, s)
	if status, _, _ := putSettings(t, s, tok, "0", settingsBody(16, `{"new":true}`)); status != http.StatusOK {
		t.Fatal("first PUT")
	}
	status, got, _ := putSettings(t, s, tok, "1", settingsBody(15, `{"old":true}`))
	if status != http.StatusConflict {
		t.Fatalf("older version: %d, want 409", status)
	}
	if got.Version != 16 || got.Revision != 1 || string(got.Settings) != `{"new":true}` {
		t.Errorf("409 body = %+v, want the current copy", got)
	}
	if !strings.Contains(got.Error, "reload") {
		t.Errorf("409 error %q should tell the player to reload", got.Error)
	}
}

func TestMeSettingsValidatesTheWrite(t *testing.T) {
	s := newMyGamesStack(t)
	tok, _ := signedInSettingsUser(t, s)

	cases := []struct {
		name    string
		ifMatch string
		body    string
		want    int
	}{
		{"no If-Match", "", settingsBody(15, `{}`), http.StatusPreconditionRequired},
		{"If-Match not a number", "abc", settingsBody(15, `{}`), http.StatusBadRequest},
		{"negative If-Match", "-1", settingsBody(15, `{}`), http.StatusBadRequest},
		{"settings an array", "0", settingsBody(15, `[]`), http.StatusBadRequest},
		{"settings missing", "0", `{"version":15}`, http.StatusBadRequest},
		{"version 0", "0", settingsBody(0, `{}`), http.StatusBadRequest},
		{"version 1001", "0", settingsBody(1001, `{}`), http.StatusBadRequest},
		{"too deep", "0", settingsBody(15, `{"a":{"b":{"c":{"d":{}}}}}`), http.StatusBadRequest},
		{"unknown field", "0", `{"version":15,"settings":{},"extra":1}`, http.StatusBadRequest},
		{"not JSON", "0", `nope`, http.StatusBadRequest},
	}
	for _, c := range cases {
		if status, _, _ := putSettings(t, s, tok, c.ifMatch, c.body); status != c.want {
			t.Errorf("%s: %d, want %d", c.name, status, c.want)
		}
	}
	if _, got := getSettings(t, s, tok); got.Revision != 0 {
		t.Errorf("a refused write created a copy: %+v", got)
	}
}

func TestMeSettingsEnforcesTheSizeCap(t *testing.T) {
	s := newMyGamesStack(t)
	tok, _ := signedInSettingsUser(t, s)

	// {"a":"xxx…"} of exactly MaxBodyBytes is accepted.
	exact := `{"a":"` + strings.Repeat("x", usersettings.MaxBodyBytes-8) + `"}`
	if len(exact) != usersettings.MaxBodyBytes {
		t.Fatalf("test body is %d bytes", len(exact))
	}
	if status, got, _ := putSettings(t, s, tok, "0", settingsBody(15, exact)); status != http.StatusOK {
		t.Fatalf("exactly 32 KiB: %d %+v", status, got)
	}
	over := `{"a":"` + strings.Repeat("x", usersettings.MaxBodyBytes-7) + `"}`
	if status, _, _ := putSettings(t, s, tok, "1", settingsBody(15, over)); status != http.StatusRequestEntityTooLarge {
		t.Errorf("32 KiB + 1: %d, want 413", status)
	}
	// A request far past the cap is cut off at the reader.
	huge := `{"a":"` + strings.Repeat("x", 4*usersettings.MaxBodyBytes) + `"}`
	if status, _, _ := putSettings(t, s, tok, "1", settingsBody(15, huge)); status != http.StatusRequestEntityTooLarge {
		t.Errorf("128 KiB: %d, want 413", status)
	}
	if _, got := getSettings(t, s, tok); got.Revision != 1 {
		t.Errorf("an oversized write moved the revision to %d", got.Revision)
	}
}

// TestMeSettingsPutIsRateLimited: one write a second per person with a
// burst of five. The limit is per person, not per address, and it is
// spent only by a write that reached the database.
func TestMeSettingsPutIsRateLimited(t *testing.T) {
	s := newUserStack(t, nil) // NOT newMyGamesStack: the limits are real here
	tok, me := signedInSettingsUser(t, s)

	// Refused writes cost nothing.
	for i := 0; i < 10; i++ {
		if status, _, _ := putSettings(t, s, tok, "", settingsBody(15, `{}`)); status != http.StatusPreconditionRequired {
			t.Fatalf("write %d without If-Match: %d", i, status)
		}
	}

	rev := 0
	for i := 0; i < 5; i++ {
		status, got, _ := putSettings(t, s, tok, strconv.Itoa(rev), settingsBody(15, `{"n":`+strconv.Itoa(i)+`}`))
		if status != http.StatusOK {
			t.Fatalf("write %d of the burst: %d", i, status)
		}
		rev = int(got.Revision)
	}
	status, _, hdr := putSettings(t, s, tok, strconv.Itoa(rev), settingsBody(15, `{"n":99}`))
	if status != http.StatusTooManyRequests {
		t.Fatalf("sixth write in a burst: %d, want 429", status)
	}
	if hdr.Get("Retry-After") == "" {
		t.Error("429 without Retry-After")
	}
	// The same person from another session shares the bucket…
	seat := playerTokenWithUser(t, s.auth, uuid.New(), uuid.New(), me, "Alice")
	if status, _, _ := putSettings(t, s, seat, strconv.Itoa(rev), settingsBody(15, `{}`)); status != http.StatusTooManyRequests {
		t.Errorf("same person, other session: %d, want 429", status)
	}
	// …and somebody else does not.
	other := identityToken(t, s.auth, dmTarget(t, s, "d-other", "Other").ID, "d-other", "Other")
	if status, _, _ := putSettings(t, s, other, "0", settingsBody(15, `{}`)); status == http.StatusTooManyRequests {
		t.Error("another person was limited by the first one's writes")
	}
	// Reads are not limited.
	for i := 0; i < 10; i++ {
		if status, _ := getSettings(t, s, tok); status != http.StatusOK {
			t.Fatalf("GET %d: %d", i, status)
		}
	}
}
