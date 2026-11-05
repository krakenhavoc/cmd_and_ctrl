package lobby

// Tests for stats_prefs.go (ADR 0217 §2, owner answer 1): the person's
// switch, the admin's switch and default, the listing, and what GET
// /me/stats reports.

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/db"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/gamestats"
)

const statsAdminDiscordID = "111111111111111111"

type statsPrefsStack struct {
	*adminStack
	db *db.DB
}

func newStatsPrefsStack(t *testing.T) statsPrefsStack {
	t.Helper()
	t.Setenv("CMDCTRL_DEV_RELAX_RATE_LIMITS", "1")
	d := openTestDB(t, t.TempDir())
	s := newAdminStackWith(t, nil, func(c *Config) {
		c.StatsPrefs = gamestats.NewSQLPrefsStore(d)
	}, statsAdminDiscordID)
	return statsPrefsStack{adminStack: s, db: d}
}

// person signs in a player whose users row exists.
func (s statsPrefsStack) person(t *testing.T, discordID, name string) (string, uuid.UUID) {
	t.Helper()
	tok, id := s.signedIn(t, discordID, name)
	if _, err := s.db.Exec(`INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES (?, ?, 1, 1)`, id.String(), name); err != nil {
		t.Fatal(err)
	}
	return tok, id
}

func decodeInto[T any](t *testing.T, s statsPrefsStack, method, path, tok string, body any) (int, T) {
	t.Helper()
	resp := do(t, s.srv, method, path, tok, body)
	defer resp.Body.Close()
	var out T
	_ = json.NewDecoder(resp.Body).Decode(&out)
	return resp.StatusCode, out
}

func (s statsPrefsStack) prefs(t *testing.T, tok string) statsPrefsResponse {
	t.Helper()
	code, out := decodeInto[statsPrefsResponse](t, s, "GET", "/me/stats/prefs", tok, nil)
	if code != http.StatusOK {
		t.Fatalf("GET /me/stats/prefs: %d", code)
	}
	return out
}

func (s statsPrefsStack) meStatsAdvanced(t *testing.T, tok string) bool {
	t.Helper()
	code, out := decodeInto[map[string]any](t, s, "GET", "/me/stats", tok, nil)
	if code != http.StatusOK {
		t.Fatalf("GET /me/stats: %d", code)
	}
	v, ok := out["advanced"].(bool)
	if !ok {
		t.Fatalf("GET /me/stats has no advanced: %v", out)
	}
	if _, ok := out["by_deck"]; !ok {
		t.Errorf("GET /me/stats lost the basic record's fields: %v", out)
	}
	return v
}

func TestStatsPrefsOffByDefaultAndTheSwitch(t *testing.T) {
	s := newStatsPrefsStack(t)
	tok, _ := s.person(t, "222", "Bea")
	got := s.prefs(t, tok)
	want := statsPrefsResponse{Available: true, Advanced: false, Choice: "default", Default: false}
	if got != want {
		t.Fatalf("a person who never touched it: %+v, want %+v", got, want)
	}
	if s.meStatsAdvanced(t, tok) {
		t.Error("GET /me/stats says advanced before the switch")
	}
	code, put := decodeInto[statsPrefsResponse](t, s, "PUT", "/me/stats/prefs", tok, map[string]bool{"advanced": true})
	if code != http.StatusOK || !put.Advanced || put.Choice != "on" || put.SetBy != "self" {
		t.Fatalf("PUT on: %d %+v", code, put)
	}
	if again := s.prefs(t, tok); again != put {
		t.Errorf("GET after PUT = %+v, PUT answered %+v", again, put)
	}
	if !s.meStatsAdvanced(t, tok) {
		t.Error("GET /me/stats does not report advanced after the switch")
	}
	for _, bad := range []any{map[string]any{}, map[string]any{"advanced": "yes"}, map[string]any{"advanced": nil}, map[string]any{"on": true}} {
		if code := status(t, s.srv, "PUT", "/me/stats/prefs", tok, bad); code != http.StatusBadRequest {
			t.Errorf("PUT %v: %d, want 400", bad, code)
		}
	}
}

// Guests, the admin token and no session get nothing: a guest has no
// stats (owner answer 1), and /me/* is 403 never 401 for a session.
func TestStatsPrefsNeedASignedInPerson(t *testing.T) {
	s := newStatsPrefsStack(t)
	guest := s.issue(t, auth.Principal{Role: auth.RolePlayer, GameID: uuid.New(), PlayerID: uuid.New()})
	for _, method := range []string{"GET", "PUT"} {
		for name, tok := range map[string]string{"guest": guest, "admin token": s.adminToken(t)} {
			if code := status(t, s.srv, method, "/me/stats/prefs", tok, map[string]bool{"advanced": true}); code != http.StatusForbidden {
				t.Errorf("%s %s: %d, want 403", name, method, code)
			}
		}
		if code := status(t, s.srv, method, "/me/stats/prefs", "", map[string]bool{"advanced": true}); code != http.StatusUnauthorized {
			t.Errorf("no session %s: %d, want 401", method, code)
		}
	}
	var n int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM stats_prefs`).Scan(&n); err != nil || n != 0 {
		t.Errorf("stats_prefs rows = %d (%v), want none", n, err)
	}
}

// Every admin route is admin-only: a player, and an allowlisted person
// in player mode, get the 403 a non-admin gets.
func TestAdvancedStatsAdminRoutesAreAdminOnly(t *testing.T) {
	s := newStatsPrefsStack(t)
	player, target := s.person(t, "333", "Cal")
	allowlisted, _ := s.person(t, statsAdminDiscordID, "Ian")
	routes := []struct{ method, path string }{
		{"GET", "/admin/stats/advanced"},
		{"PUT", "/admin/stats/advanced-default"},
		{"PUT", "/admin/users/" + target.String() + "/advanced-stats"},
	}
	for _, r := range routes {
		for name, tok := range map[string]string{"player": player, "player mode": allowlisted} {
			if code := status(t, s.srv, r.method, r.path, tok, map[string]bool{"on": true}); code != http.StatusForbidden {
				t.Errorf("%s %s as %s: %d, want 403", r.method, r.path, name, code)
			}
		}
		if code := status(t, s.srv, r.method, r.path, "", map[string]bool{"on": true}); code != http.StatusUnauthorized {
			t.Errorf("%s %s with no session: %d, want 401", r.method, r.path, code)
		}
	}
	if s.prefs(t, player).Choice != "default" {
		t.Error("a refused admin call changed the person's switch")
	}
}

// The admin's switch is recorded as set by an admin, is audited with
// who did it, and the person can still switch it off.
func TestAdminSwitchesAdvancedStatsForAPerson(t *testing.T) {
	s := newStatsPrefsStack(t)
	ptok, target := s.person(t, "444", "Dee")
	atok, admin := s.person(t, statsAdminDiscordID, "Ian")
	s.adminModeOn(t, admin)

	code, out := decodeInto[adminAccountStats](t, s, "PUT", "/admin/users/"+target.String()+"/advanced-stats", atok, map[string]bool{"on": true})
	if code != http.StatusOK || out.Choice != "on" || out.SetBy != "admin" || !out.Advanced || out.UserID != target.String() {
		t.Fatalf("admin switch on: %d %+v", code, out)
	}
	var audited bool
	for _, line := range strings.Split(s.log.String(), "\n") {
		if strings.Contains(line, "admin action") && strings.Contains(line, "advanced stats set") &&
			strings.Contains(line, "admin_user_id="+admin.String()) && strings.Contains(line, "user_id="+target.String()) &&
			strings.Contains(line, "on=true") {
			audited = true
		}
	}
	if !audited {
		t.Errorf("no audit line naming the admin, the person and the switch:\n%s", s.log.String())
	}
	if got := s.prefs(t, ptok); !got.Advanced || got.SetBy != "admin" {
		t.Errorf("the person sees %+v, want on and set by an admin", got)
	}
	if !s.meStatsAdvanced(t, ptok) {
		t.Error("GET /me/stats does not report the admin's switch")
	}
	// The person switches it off again.
	if _, off := decodeInto[statsPrefsResponse](t, s, "PUT", "/me/stats/prefs", ptok, map[string]bool{"advanced": false}); off.Advanced || off.SetBy != "self" {
		t.Errorf("the person's switch off: %+v", off)
	}

	// The listing shows the account with who set it.
	code, list := decodeInto[adminAdvancedStatsResponse](t, s, "GET", "/admin/stats/advanced", atok, nil)
	if code != http.StatusOK || len(list.Accounts) != 1 {
		t.Fatalf("listing: %d %+v", code, list)
	}
	if a := list.Accounts[0]; a.UserID != target.String() || a.Choice != "off" || a.SetBy != "self" || a.Advanced || a.UpdatedAt == 0 {
		t.Errorf("listing row = %+v", a)
	}
	if list.Default.Advanced || list.Default.Stored {
		t.Errorf("listing default = %+v, want off and not stored", list.Default)
	}

	for _, bad := range []struct {
		path string
		body any
		want int
	}{
		{"/admin/users/not-a-uuid/advanced-stats", map[string]bool{"on": true}, http.StatusBadRequest},
		{"/admin/users/" + target.String() + "/advanced-stats", map[string]any{}, http.StatusBadRequest},
		{"/admin/users/" + uuid.NewString() + "/advanced-stats", map[string]bool{"on": true}, http.StatusNotFound},
	} {
		if code := status(t, s.srv, "PUT", bad.path, atok, bad.body); code != bad.want {
			t.Errorf("PUT %s %v: %d, want %d", bad.path, bad.body, code, bad.want)
		}
	}
}

// The default is stored on the server and audited. It reaches everyone
// who never touched the switch, and anyone who switched it off
// themselves stays off when it flips.
func TestAdvancedStatsDefaultFallbackAndSelfOffStaysOff(t *testing.T) {
	s := newStatsPrefsStack(t)
	untouched, _ := s.person(t, "555", "Eve")
	selfOff, _ := s.person(t, "666", "Fin")
	tok := s.adminToken(t)
	if code := status(t, s.srv, "PUT", "/me/stats/prefs", selfOff, map[string]bool{"advanced": false}); code != http.StatusOK {
		t.Fatalf("self off: %d", code)
	}
	if s.prefs(t, untouched).Advanced || s.prefs(t, selfOff).Advanced {
		t.Fatal("advanced is on before any default was stored")
	}

	code, def := decodeInto[adminDefaultResponse](t, s, "PUT", "/admin/stats/advanced-default", tok, map[string]bool{"on": true})
	if code != http.StatusOK || !def.Advanced || !def.Stored || def.UpdatedAt == 0 {
		t.Fatalf("PUT default: %d %+v", code, def)
	}
	if !strings.Contains(s.log.String(), "advanced stats default set") {
		t.Errorf("no audit line for the default:\n%s", s.log.String())
	}
	if got := s.prefs(t, untouched); !got.Advanced || got.Choice != "default" || !got.Default {
		t.Errorf("an untouched account after the default flips: %+v, want on, following the default", got)
	}
	if !s.meStatsAdvanced(t, untouched) {
		t.Error("GET /me/stats does not follow the stored default")
	}
	if got := s.prefs(t, selfOff); got.Advanced || got.Choice != "off" {
		t.Errorf("someone who switched it off after the default flips: %+v, want off", got)
	}
	if s.meStatsAdvanced(t, selfOff) {
		t.Error("GET /me/stats turned on advanced for someone who switched it off")
	}
	_, list := decodeInto[adminAdvancedStatsResponse](t, s, "GET", "/admin/stats/advanced", tok, nil)
	if !list.Default.Advanced || !list.Default.Stored {
		t.Errorf("listing default = %+v", list.Default)
	}

	// Flipping it back reaches the untouched account again.
	status(t, s.srv, "PUT", "/admin/stats/advanced-default", tok, map[string]bool{"on": false})
	if s.prefs(t, untouched).Advanced {
		t.Error("the default switched back off did not reach the untouched account")
	}
	if code := status(t, s.srv, "PUT", "/admin/stats/advanced-default", tok, map[string]any{"on": "yes"}); code != http.StatusBadRequest {
		t.Errorf("a bad body: %d, want 400", code)
	}
}

// With no database everyone follows gamestats.AdvancedDefault, the
// reads still answer, and the writes are 503.
func TestStatsPrefsWithNoDatabase(t *testing.T) {
	s := newAdminStack(t, statsAdminDiscordID)
	tok, _ := s.signedIn(t, "777", "Gus")
	resp := do(t, s.srv, "GET", "/me/stats/prefs", tok, nil)
	var got statsPrefsResponse
	_ = json.NewDecoder(resp.Body).Decode(&got)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK || got.Available || got.Advanced != gamestats.AdvancedDefault || got.Choice != "default" {
		t.Errorf("GET with no database: %d %+v", resp.StatusCode, got)
	}
	if code := status(t, s.srv, "PUT", "/me/stats/prefs", tok, map[string]bool{"advanced": true}); code != http.StatusServiceUnavailable {
		t.Errorf("PUT with no database: %d, want 503", code)
	}
	admin := s.adminToken(t)
	for _, r := range []struct{ method, path string }{
		{"GET", "/admin/stats/advanced"},
		{"PUT", "/admin/stats/advanced-default"},
		{"PUT", "/admin/users/" + uuid.NewString() + "/advanced-stats"},
	} {
		if code := status(t, s.srv, r.method, r.path, admin, map[string]bool{"on": true}); code != http.StatusServiceUnavailable {
			t.Errorf("%s %s with no database: %d, want 503", r.method, r.path, code)
		}
	}
}
