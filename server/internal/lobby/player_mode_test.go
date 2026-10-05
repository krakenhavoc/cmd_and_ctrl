package lobby

// player_mode_test.go covers ADR 0112 §2 (Delivery PR 1): player mode.
// An allowlisted person is an admin only in admin mode, which is off
// by default, switched by PUT /me/admin-mode, lapses after 12 hours,
// and ends on sign-out-everywhere. In player mode every admin call
// site answers exactly as it does for a non-admin signed-in person, and
// the table at the bottom of this file asks every one of them.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deckrequests"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/users"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// --- PUT /me/admin-mode -----------------------------------------------

type adminModeAnswer struct {
	Admin           bool   `json:"admin"`
	AdminMode       bool   `json:"admin_mode"`
	AdminModeEndsAt *int64 `json:"admin_mode_ends_at"`
	Error           string `json:"error"`
}

func putMode(t *testing.T, s *adminStack, tok string, body any) (int, adminModeAnswer) {
	t.Helper()
	resp := do(t, s.srv, http.MethodPut, "/me/admin-mode", tok, body)
	defer resp.Body.Close()
	var out adminModeAnswer
	_ = json.NewDecoder(resp.Body).Decode(&out)
	return resp.StatusCode, out
}

func meOf(t *testing.T, s *adminStack, tok string) meResponse {
	t.Helper()
	resp := doGet(t, s.srv, "/me", tok)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("/me: %d", resp.StatusCode)
	}
	var me meResponse
	if err := json.NewDecoder(resp.Body).Decode(&me); err != nil {
		t.Fatal(err)
	}
	return me
}

// An allowlisted person starts in player mode, switches admin mode on
// and off, and every session they hold follows at its next request.
func TestAdminModeSwitchOnAndOff(t *testing.T) {
	t.Setenv("CMDCTRL_DEV_RELAX_RATE_LIMITS", "1")
	s := newAdminStack(t, listedDiscordID)
	at := time.Date(2026, 10, 2, 20, 0, 0, 0, time.UTC)
	s.modes.SetClock(func() time.Time { return at })
	laptop, user := s.signedIn(t, listedDiscordID, "Owner")
	// A second session of the same person, minted before the switch.
	phone := s.issue(t, auth.Principal{Role: auth.RoleIdentified, UserID: user, DiscordID: listedDiscordID})
	other, _ := s.lobby.Create("someone else's table")
	archive := "/games/" + other.ID.String() + "/archive"

	me := meOf(t, s, laptop)
	if me.Admin || me.AdminMode || !me.AdminAllowed || me.AdminModeEndsAt != 0 {
		t.Fatalf("/me before any switch = admin %v mode %v allowed %v ends %d; want player mode, allowed", me.Admin, me.AdminMode, me.AdminAllowed, me.AdminModeEndsAt)
	}
	if got := status(t, s.srv, "POST", archive, laptop, nil); got != http.StatusForbidden {
		t.Fatalf("an admin route in player mode: %d, want 403", got)
	}

	code, out := putMode(t, s, laptop, map[string]bool{"on": true})
	if code != http.StatusOK || !out.Admin || !out.AdminMode || out.AdminModeEndsAt == nil {
		t.Fatalf("switch on: %d %+v", code, out)
	}
	if want := at.Add(12 * time.Hour).UnixMilli(); *out.AdminModeEndsAt != want {
		t.Errorf("admin_mode_ends_at = %d, want %d (12 hours)", *out.AdminModeEndsAt, want)
	}
	// Both sessions are admins now: nothing in either token changed.
	for name, tok := range map[string]string{"laptop": laptop, "phone": phone} {
		me := meOf(t, s, tok)
		if !me.Admin || !me.AdminMode || me.AdminModeEndsAt != at.Add(12*time.Hour).UnixMilli() {
			t.Errorf("%s /me after switching on = %+v", name, me)
		}
	}
	if got := status(t, s.srv, "POST", archive, phone, nil); got != http.StatusOK {
		t.Errorf("an admin route in admin mode, from the other session: %d, want 200", got)
	}

	// On while on restarts the 12 hours.
	s.modes.SetClock(func() time.Time { return at.Add(3 * time.Hour) })
	code, out = putMode(t, s, phone, map[string]bool{"on": true})
	if code != http.StatusOK || out.AdminModeEndsAt == nil || *out.AdminModeEndsAt != at.Add(15*time.Hour).UnixMilli() {
		t.Errorf("switch on again: %d %+v, want the 12 hours restarted", code, out)
	}

	code, out = putMode(t, s, laptop, map[string]bool{"on": false})
	if code != http.StatusOK || out.Admin || out.AdminMode || out.AdminModeEndsAt != nil {
		t.Fatalf("switch off: %d %+v", code, out)
	}
	if got := status(t, s.srv, "DELETE", archive, phone, nil); got != http.StatusForbidden {
		t.Errorf("an admin route after switching off: %d, want 403", got)
	}
	// Off while off is a 200 no-op.
	if code, out := putMode(t, s, laptop, map[string]bool{"on": false}); code != http.StatusOK || out.AdminMode {
		t.Errorf("switch off twice: %d %+v", code, out)
	}

	// Audit: every switch, at Info, with the user and never a token.
	logs := s.log.String()
	if strings.Count(logs, `msg="admin mode on"`) != 2 || strings.Count(logs, `msg="admin mode off"`) != 2 {
		t.Errorf("want two on and two off audit lines:\n%s", logs)
	}
	if !strings.Contains(logs, "admin_user_id="+user.String()) {
		t.Error("the switch was not audited with admin_user_id")
	}
	for _, tok := range []string{laptop, phone} {
		if strings.Contains(logs, tok) {
			t.Error("the audit log carries a session token")
		}
	}
}

// Only an allowlisted person may switch. Everyone else, the shared
// token included, is one 403, "not an admin".
func TestAdminModeSwitchIsRefusedToEveryoneElse(t *testing.T) {
	t.Setenv("CMDCTRL_DEV_RELAX_RATE_LIMITS", "1")
	s := newAdminStack(t, listedDiscordID)
	unlisted, _ := s.signedIn(t, unlistedDiscordID, "Stranger")
	meta, _ := s.lobby.Create("t")
	_, guestPlayer, _ := s.lobby.Join(meta.ID, meta.InviteToken, "Guest")
	guest := s.issue(t, auth.Principal{Role: auth.RolePlayer, GameID: meta.ID, PlayerID: guestPlayer, Name: "Guest"})
	reclaim := s.issue(t, auth.Principal{Role: auth.RolePlayer, GameID: meta.ID, PlayerID: guestPlayer, DiscordID: listedDiscordID})

	for name, tok := range map[string]string{
		"shared token":       s.adminToken(t),
		"unlisted signed-in": unlisted,
		"guest seat":         guest,
		"reclaim ticket with a listed Discord ID": reclaim,
	} {
		code, out := putMode(t, s, tok, map[string]bool{"on": true})
		if code != http.StatusForbidden || out.Error != "not an admin" {
			t.Errorf("%s: %d %q, want 403 \"not an admin\"", name, code, out.Error)
		}
	}
	if code, _ := putMode(t, s, "", map[string]bool{"on": true}); code != http.StatusUnauthorized {
		t.Errorf("no credential: %d, want 401", code)
	}

	listed, user := s.signedIn(t, listedDiscordID, "Owner")
	for name, body := range map[string]any{"empty object": map[string]any{}, "a string": map[string]any{"on": "yes"}, "an unknown field": map[string]any{"on": true, "until": 5}} {
		if code, _ := putMode(t, s, listed, body); code != http.StatusBadRequest {
			t.Errorf("%s: %d, want 400", name, code)
		}
	}
	if s.modes.On(user, time.Now()) {
		t.Error("a refused request switched admin mode on")
	}
}

// A server with no user database has no modes: the switch is a 503 and
// nobody on the list is an admin.
func TestAdminModeWithNoUserDatabase(t *testing.T) {
	s := newAdminStackWith(t, nil, func(c *Config) {
		c.Admins = NewAdmins(NewAdminList(listedDiscordID), nil)
	}, listedDiscordID)
	listed, _ := s.signedIn(t, listedDiscordID, "Owner")
	if code, _ := putMode(t, s, listed, map[string]bool{"on": true}); code != http.StatusServiceUnavailable {
		t.Errorf("switch with no modes: %d, want 503", code)
	}
	if me := meOf(t, s, listed); me.Admin || me.AdminMode || !me.AdminAllowed {
		t.Errorf("/me with no modes = %+v, want allowed but in player mode", me)
	}
}

// One switch every 2 seconds per person, with a burst of 5.
func TestAdminModeSwitchIsRateLimited(t *testing.T) {
	s := newAdminStack(t, listedDiscordID)
	listed, _ := s.signedIn(t, listedDiscordID, "Owner")
	for i := 0; i < 5; i++ {
		if code, _ := putMode(t, s, listed, map[string]bool{"on": i%2 == 0}); code != http.StatusOK {
			t.Fatalf("switch %d: %d", i+1, code)
		}
	}
	if code, _ := putMode(t, s, listed, map[string]bool{"on": true}); code != http.StatusTooManyRequests {
		t.Errorf("sixth switch in a burst: %d, want 429", code)
	}
	// Another person has their own bucket.
	other, _ := s.signedIn(t, listedDiscordID, "Owner's other account")
	if code, _ := putMode(t, s, other, map[string]bool{"on": true}); code != http.StatusOK {
		t.Errorf("another person's first switch: %d, want 200", code)
	}
}

// A failed write fails closed: switching on stays off, switching off
// is off in this process. Both are a 500.
func TestAdminModeSwitchFailsClosed(t *testing.T) {
	t.Setenv("CMDCTRL_DEV_RELAX_RATE_LIMITS", "1")
	store := &failingModeStore{memAdminModeStore: memAdminModeStore{rows: map[uuid.UUID]time.Time{}}}
	modes, err := users.NewAdminModes(context.Background(), store)
	if err != nil {
		t.Fatal(err)
	}
	s := newAdminStackWith(t, nil, func(c *Config) {
		c.Admins = NewAdmins(NewAdminList(listedDiscordID), modes)
	}, listedDiscordID)
	s.modes = modes
	listed, user := s.signedIn(t, listedDiscordID, "Owner")

	store.fail.Store(true)
	if code, _ := putMode(t, s, listed, map[string]bool{"on": true}); code != http.StatusInternalServerError {
		t.Errorf("a failed switch on: %d, want 500", code)
	}
	if meOf(t, s, listed).Admin {
		t.Error("a failed switch on left the person an admin")
	}

	store.fail.Store(false)
	if code, _ := putMode(t, s, listed, map[string]bool{"on": true}); code != http.StatusOK {
		t.Fatal("switch on")
	}
	store.fail.Store(true)
	if code, _ := putMode(t, s, listed, map[string]bool{"on": false}); code != http.StatusInternalServerError {
		t.Errorf("a failed switch off: %d, want 500", code)
	}
	if meOf(t, s, listed).Admin || modes.On(user, time.Now()) {
		t.Error("a failed switch off left the person an admin")
	}
}

type failingModeStore struct {
	memAdminModeStore
	fail atomic.Bool
}

func (f *failingModeStore) SetAdminMode(ctx context.Context, id uuid.UUID, at time.Time) error {
	if f.fail.Load() {
		return errors.New("disk full")
	}
	return f.memAdminModeStore.SetAdminMode(ctx, id, at)
}

// --- the sockets: close 4001 and bind again ----------------------------

// readClose reads c until it closes and returns the close error.
func readClose(t *testing.T, c *websocket.Conn) *websocket.CloseError {
	t.Helper()
	_ = c.SetReadDeadline(time.Now().Add(3 * time.Second))
	for {
		_, _, err := c.ReadMessage()
		if err == nil {
			continue
		}
		var ce *websocket.CloseError
		if errors.As(err, &ce) {
			return ce
		}
		t.Fatalf("socket ended without a close frame: %v", err)
		return nil
	}
}

func dialWS(t *testing.T, s *adminStack, tok, query string) (*websocket.Conn, int) {
	t.Helper()
	u := "ws" + strings.TrimPrefix(s.srv.URL, "http") + "/ws?" + query
	h := http.Header{}
	h.Set("Authorization", "Bearer "+tok)
	conn, resp, err := websocket.DefaultDialer.Dial(u, h)
	if resp != nil {
		resp.Body.Close()
	}
	if err != nil {
		if resp == nil {
			t.Fatalf("dial %s: %v", query, err)
		}
		return nil, resp.StatusCode
	}
	t.Cleanup(func() { _ = conn.Close() })
	return conn, resp.StatusCode
}

func waitForCount(t *testing.T, hub *ws.Hub, n int) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for hub.Count() != n && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if got := hub.Count(); got != n {
		t.Fatalf("hub has %d sockets, want %d", got, n)
	}
}

// A switch closes the sockets it made wrong with 4001, and the
// reconnect binds with the new answer (ADR 0112 §2 item 5). In player
// mode the person gets their own seat and nothing else.
func TestAdminModeSwitchRebindsTheSockets(t *testing.T) {
	t.Setenv("CMDCTRL_DEV_RELAX_RATE_LIMITS", "1")
	s := newAdminStack(t, listedDiscordID)
	mine, _ := s.lobby.Create("mine")
	theirs, _ := s.lobby.Create("theirs")
	_, seat, err := s.lobby.Join(mine.ID, mine.InviteToken, "Owner")
	if err != nil {
		t.Fatal(err)
	}
	user := uuid.New()
	seatTok := s.issue(t, auth.Principal{Role: auth.RolePlayer, UserID: user, GameID: mine.ID, PlayerID: seat, DiscordID: listedDiscordID})

	// Player mode: their own seat, never another table.
	own, code := dialWS(t, s, seatTok, "game="+mine.ID.String())
	if code != http.StatusSwitchingProtocols {
		t.Fatalf("own seat in player mode: %d", code)
	}
	if _, code := dialWS(t, s, seatTok, "game="+theirs.ID.String()); code != http.StatusForbidden {
		t.Errorf("another table in player mode: %d, want 403", code)
	}
	// The shared token's socket: no user, no mode, never closed by a
	// person's switch.
	token, _ := dialWS(t, s, s.adminToken(t), "game="+theirs.ID.String())
	waitForCount(t, s.hub, 2)

	// Switch on: the ordinary socket is now wrong, and closes with 4001.
	if code, _ := putMode(t, s, seatTok, map[string]bool{"on": true}); code != http.StatusOK {
		t.Fatalf("switch on: %d", code)
	}
	if ce := readClose(t, own); ce.Code != ws.AdminModeChangedCode || ce.Text != ws.AdminModeChangedReason {
		t.Errorf("own seat after switching on: close %d %q, want 4001", ce.Code, ce.Text)
	}
	b, err := s.wsAuth.AuthorizeUpgrade(upgradeRequest(seatTok, "game="+mine.ID.String()))
	if err != nil || !b.Admin || b.PlayerID != seat {
		t.Errorf("reconnect in admin mode: %+v %v; want their seat with Binding.Admin", b, err)
	}
	elsewhere, code := dialWS(t, s, seatTok, "game="+theirs.ID.String())
	if code != http.StatusSwitchingProtocols {
		t.Fatalf("another table in admin mode: %d", code)
	}
	again, _ := dialWS(t, s, seatTok, "game="+mine.ID.String())
	waitForCount(t, s.hub, 3)

	// Switch off: both admin sockets close; the token's stays up.
	if code, _ := putMode(t, s, seatTok, map[string]bool{"on": false}); code != http.StatusOK {
		t.Fatalf("switch off: %d", code)
	}
	for name, c := range map[string]*websocket.Conn{"another table": elsewhere, "own seat": again} {
		if ce := readClose(t, c); ce.Code != ws.AdminModeChangedCode {
			t.Errorf("%s after switching off: close %d, want 4001", name, ce.Code)
		}
	}
	waitForCount(t, s.hub, 1)
	if _, code := dialWS(t, s, seatTok, "game="+theirs.ID.String()); code != http.StatusForbidden {
		t.Errorf("another table after switching off: %d, want 403", code)
	}
	b, err = s.wsAuth.AuthorizeUpgrade(upgradeRequest(seatTok, "game="+mine.ID.String()))
	if err != nil || b.Admin || b.PlayerID != seat {
		t.Errorf("reconnect in player mode: %+v %v; want their own seat, no Binding.Admin", b, err)
	}
	_ = token.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	if _, _, err := token.ReadMessage(); err != nil {
		var ce *websocket.CloseError
		if errors.As(err, &ce) {
			t.Errorf("the shared token's socket was closed by a person's switch: %v", ce)
		}
	}
}

// The sweeper ends a lapsed admin mode, closes the person's admin
// sockets with 4001 and logs "admin mode lapsed" (owner answer 1).
func TestAdminModeSweeperEndsALapse(t *testing.T) {
	s := newAdminStack(t, listedDiscordID)
	theirs, _ := s.lobby.Create("theirs")
	at := time.Date(2026, 10, 2, 20, 0, 0, 0, time.UTC)
	now := at
	var mu sync.Mutex
	s.modes.SetClock(func() time.Time { mu.Lock(); defer mu.Unlock(); return now })
	tok, user := s.signedIn(t, listedDiscordID, "Owner")
	s.adminModeOn(t, user)
	conn, code := dialWS(t, s, tok, "game="+theirs.ID.String())
	if code != http.StatusSwitchingProtocols {
		t.Fatalf("admin socket: %d", code)
	}
	waitForCount(t, s.hub, 1)

	// Before the 12 hours: the sweep does nothing.
	sweepAdminModes(context.Background(), s.cfg.Admins, s.hub, s.cfg.Log)
	if !meOf(t, s, tok).Admin {
		t.Fatal("swept before the 12 hours")
	}

	mu.Lock()
	now = at.Add(12 * time.Hour)
	mu.Unlock()
	// HTTP sees the lapse at once, with no sweep.
	if me := meOf(t, s, tok); me.Admin || me.AdminMode {
		t.Errorf("/me at 12 hours = %+v, want player mode", me)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		RunAdminModeSweeper(ctx, s.cfg.Admins, s.hub, s.cfg.Log, time.Hour)
		close(done)
	}()
	// The first sweep runs at start.
	if ce := readClose(t, conn); ce.Code != ws.AdminModeChangedCode {
		t.Errorf("admin socket after the lapse: close %d, want 4001", ce.Code)
	}
	cancel()
	<-done
	if !strings.Contains(s.log.String(), `msg="admin mode lapsed" admin_user_id=`+user.String()) {
		t.Errorf("no lapse audit line:\n%s", s.log.String())
	}
	if _, code := dialWS(t, s, tok, "game="+theirs.ID.String()); code != http.StatusForbidden {
		t.Errorf("another table after the lapse: %d, want 403", code)
	}
}

// Signing out everywhere ends admin mode, and so does the admin's
// revoke (ADR 0112 §2 item 7). Over the real database: one statement
// moves the watermark and zeroes admin_mode_at.
func TestSignOutEverywhereEndsAdminMode(t *testing.T) {
	var modes *users.AdminModes
	var us *users.SQLStore
	s := newRevocationStack(t, func(c *Config) {
		us = c.Users.(*users.SQLStore)
		var err error
		modes, err = users.NewAdminModes(context.Background(), us)
		if err != nil {
			t.Fatal(err)
		}
		c.Revocations.(*users.Revocations).EndAdminModeOnRevoke(modes)
		c.Admins = NewAdmins(NewAdminList("alice", "bob"), modes)
	})
	adminStatus := func(tok string) (bool, bool) {
		resp := doGet(t, s.srv, "/me", tok)
		defer resp.Body.Close()
		var me meResponse
		_ = json.NewDecoder(resp.Body).Decode(&me)
		return me.Admin, me.AdminMode
	}
	switchOn := func(tok string) {
		resp := do(t, s.srv, http.MethodPut, "/me/admin-mode", tok, map[string]bool{"on": true})
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("switch on: %d", resp.StatusCode)
		}
	}

	alice := signInAs(t, s, "alice")
	switchOn(alice)
	if admin, _ := adminStatus(alice); !admin {
		t.Fatal("alice is not an admin after switching on")
	}
	resp := post(t, s.srv, "/logout/everywhere", alice)
	resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("logout everywhere: %d", resp.StatusCode)
	}
	afterAMillisecond()
	alice = signInAs(t, s, "alice")
	if admin, mode := adminStatus(alice); admin || mode {
		t.Error("admin mode survived logout-everywhere")
	}
	reloaded, _ := users.NewAdminModes(context.Background(), us)
	if reloaded.On(mustValidate(t, s.auth, alice).UserID, time.Now()) {
		t.Error("admin mode survived logout-everywhere in the database")
	}

	// The admin's revoke of someone else does the same.
	bob := signInAs(t, s, "bob")
	switchOn(bob)
	bobUser := mustValidate(t, s.auth, bob).UserID
	resp = post(t, s.srv, "/admin/users/"+bobUser.String()+"/revoke-sessions", adminToken(t, s.srv))
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("admin revoke: %d", resp.StatusCode)
	}
	if modes.On(bobUser, time.Now()) {
		t.Error("admin mode survived the admin's revoke")
	}
}

// --- the guard's second rule -------------------------------------------

// adminBypassOffences finds every call that asks the allowlist or the
// mode directly, AdminList.Has or AdminModes.On, outside admins.go.
//
// It works on syntax, so it is deliberately wide: a file is in scope
// when it is in package lobby or users, or imports either (only those
// can name the two types), and every `.Has` or `.On` selector in it is
// an offence, called or taken as a method value. Interfaces anywhere in
// the server that declare a Has or On method are offences too, since
// one would let a package that imports neither call the methods
// through it. A false positive costs a rename; a false negative would
// cost a route that skips the mode.
func adminBypassOffences(fset *token.FileSet, path string, f *ast.File) []string {
	rel := filepath.ToSlash(path)
	if strings.HasSuffix(rel, "internal/lobby/admins.go") {
		return nil
	}
	inScope := f.Name.Name == "lobby" || f.Name.Name == "users"
	for _, imp := range f.Imports {
		p := strings.Trim(imp.Path.Value, `"`)
		if strings.HasSuffix(p, "/internal/lobby") || strings.HasSuffix(p, "/internal/users") {
			inScope = true
		}
	}
	var out []string
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.SelectorExpr:
			if inScope && (x.Sel.Name == "Has" || x.Sel.Name == "On") {
				out = append(out, fmt.Sprintf("%s: .%s", fset.Position(x.Pos()), x.Sel.Name))
			}
		case *ast.InterfaceType:
			for _, m := range x.Methods.List {
				for _, name := range m.Names {
					if name.Name == "Has" || name.Name == "On" {
						out = append(out, fmt.Sprintf("%s: interface method %s", fset.Position(name.Pos()), name.Name))
					}
				}
			}
		}
		return true
	})
	return out
}

// TestAllowlistAndModeAreAskedOnlyInAdminsGo is the guard's second rule
// (ADR 0112 §2 item 2): AdminList.Has and AdminModes.On are called only
// inside admins.go, so no route can test the allowlist and skip the
// mode. Every admin decision goes through isAdmin, isAdminPrincipal or
// isAllowlisted.
func TestAllowlistAndModeAreAskedOnlyInAdminsGo(t *testing.T) {
	root := filepath.Join("..", "..")
	fset := token.NewFileSet()
	var offences []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			switch d.Name() {
			case "testdata", "vendor", "node_modules":
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		f, perr := parser.ParseFile(fset, path, nil, 0)
		if perr != nil {
			return perr
		}
		offences = append(offences, adminBypassOffences(fset, path, f)...)
		return nil
	})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}
	for _, o := range offences {
		t.Errorf("%s: ask isAdmin(p), isAdminPrincipal or isAllowlisted in admins.go instead; a direct allowlist or mode check can skip the other half (ADR 0112 §2 item 2)", o)
	}
}

// keptServerCredentialSites are the functions that ask
// isServerCredential, "is this the shared token": the bot's own paths,
// which a person never takes in either mode, so player mode leaves them
// as they are (ADR 0112 §2, "Kept on purpose"). Each count is how many
// times the function asks.
var keptServerCredentialSites = map[string]int{
	"deckCoverageLimit":  1, // the bot's own coverage bucket
	"deckRequesterFor":   1, // filing a deck request in a named member's name
	"deckRequestIPLimit": 1, // the bot's deck request spends the IP bucket in the handler, for its named member (#2052)
	"file":               1, // file: the issue's "filed by the bot" flag
	"request":            1, // deck_id on POST /deck-requests is a person's library; the token is refused (ADR 0112 PR 2, #2001)
	"callerKey":          1, // the per-caller limits' shared admin bucket
	"createGameWith":     1, // the token creates tables with no creator and no open-table cap
	"practiceOwner":      1, // the token's practice table
	"inviteDM":           1, // naming a raw Discord snowflake
	"AuthorizeUpgrade":   1, // the token never has an own binding
}

// TestServerCredentialSitesAreAllKept: every isServerCredential site is
// one of the kept bot paths above. A new one fails here until it is
// classified: either it is the token's own path and joins the list, or
// it is an admin decision and must ask isAdmin instead, which puts it in
// the same-answer table.
func TestServerCredentialSitesAreAllKept(t *testing.T) {
	fset := token.NewFileSet()
	files, err := parseNonTestGoFiles(fset, ".")
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]int{}
	for _, f := range files {
		for _, decl := range f.Decls {
			fd, ok := decl.(*ast.FuncDecl)
			if !ok || fd.Body == nil || fd.Name.Name == "isAdminPrincipal" {
				continue
			}
			ast.Inspect(fd.Body, func(n ast.Node) bool {
				if call, ok := n.(*ast.CallExpr); ok {
					if id, ok := call.Fun.(*ast.Ident); ok && id.Name == "isServerCredential" {
						got[fd.Name.Name]++
					}
				}
				return true
			})
		}
	}
	for fn, n := range got {
		if keptServerCredentialSites[fn] != n {
			t.Errorf("%s asks isServerCredential %d times; keptServerCredentialSites says %d. Classify it: a path only the shared token takes joins the list, an admin decision asks isAdmin (ADR 0112 §2)", fn, n, keptServerCredentialSites[fn])
		}
	}
	for fn, n := range keptServerCredentialSites {
		if got[fn] == 0 {
			t.Errorf("keptServerCredentialSites lists %s (%d), which no longer asks isServerCredential", fn, n)
		}
	}
}

// isAllowlisted answers "may this person switch admin mode on", never
// "is this an admin". Only the predicate, the switch and /me ask it.
func TestIsAllowlistedIsAskedOnlyByTheSwitchAndMe(t *testing.T) {
	allowed := map[string]bool{"isAdminPrincipal": true, "adminModeOf": true, "putAdminMode": true, "me": true}
	fset := token.NewFileSet()
	files, err := parseNonTestGoFiles(fset, ".")
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range files {
		for _, decl := range f.Decls {
			name := ""
			if fd, ok := decl.(*ast.FuncDecl); ok {
				name = fd.Name.Name
			}
			ast.Inspect(decl, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}
				if id, ok := call.Fun.(*ast.Ident); ok && id.Name == "isAllowlisted" && !allowed[name] {
					t.Errorf("%s: isAllowlisted in %s; an admin decision asks isAdmin(p), which also checks the mode (ADR 0112 §2 item 2)", fset.Position(call.Pos()), name)
				}
				return true
			})
		}
	}
}

// The rule catches the bypasses it exists for, and not a clean file.
func TestAdminBypassRuleCatchesABypass(t *testing.T) {
	cases := map[string]struct {
		src  string
		want int
	}{
		"allowlist only": {`package lobby
func sneaky(c Config, p auth.Principal) bool { return c.Admins.list.Has(p.DiscordID) }`, 1},
		"mode only, from another package": {`package bot
import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/users"
func sneaky(m *users.AdminModes, id uuid.UUID) bool { return m.On(id, time.Now()) }`, 1},
		"a method value": {`package lobby
func sneaky(l *AdminList) func(string) bool { return l.Has }`, 1},
		"an interface to call it through": {`package ws
type allowlist interface{ Has(string) bool }`, 1},
		"clean": {`package lobby
func fine(c Config, p auth.Principal) bool { return c.isAdmin(p) }`, 0},
		"out of scope": {`package game
func fine(c *Card) bool { return c.Unlocked.Has(1) }`, 0},
	}
	for name, tc := range cases {
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, name+".go", tc.src, 0)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if got := adminBypassOffences(fset, "internal/x/"+name+".go", f); len(got) != tc.want {
			t.Errorf("%s: %d offences %v, want %d", name, len(got), got, tc.want)
		}
	}
}

// --- every admin call site answers the same in player mode -------------

// adminCallSites lists, by syntax, every admin decision in the lobby
// package's non-test code: each function that calls isAdmin,
// isAdminPrincipal or requireAdmin (by name), and, inside Handler, each
// route whose handler goes through requireAdmin. The definitions
// themselves are not sites.
func adminCallSites(t *testing.T) map[string]bool {
	t.Helper()
	fset := token.NewFileSet()
	files, err := parseNonTestGoFiles(fset, ".")
	if err != nil {
		t.Fatal(err)
	}
	definitions := map[string]bool{"isAdmin": true, "isAdminPrincipal": true, "requireAdmin": true}
	isAdminCall := func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return false
		}
		switch fn := call.Fun.(type) {
		case *ast.SelectorExpr:
			return fn.Sel.Name == "isAdmin"
		case *ast.Ident:
			return fn.Name == "isAdminPrincipal" || fn.Name == "requireAdmin"
		}
		return false
	}
	sites := map[string]bool{}
	for _, f := range files {
		for _, decl := range f.Decls {
			fd, ok := decl.(*ast.FuncDecl)
			if !ok || fd.Body == nil || definitions[fd.Name.Name] {
				continue
			}
			if fd.Name.Name == "Handler" {
				ast.Inspect(fd.Body, func(n ast.Node) bool {
					call, ok := n.(*ast.CallExpr)
					if !ok || len(call.Args) != 2 {
						return true
					}
					sel, ok := call.Fun.(*ast.SelectorExpr)
					if !ok || sel.Sel.Name != "Handle" {
						return true
					}
					admin := false
					ast.Inspect(call.Args[1], func(n ast.Node) bool {
						if isAdminCall(n) {
							admin = true
						}
						return true
					})
					if !admin {
						return true
					}
					lit, ok := call.Args[0].(*ast.BasicLit)
					if !ok {
						t.Errorf("an admin route in Handler with a non-literal pattern at %s", fset.Position(call.Pos()))
						return true
					}
					pattern, _ := strconv.Unquote(lit.Value)
					sites["route:"+pattern] = true
					return true
				})
				continue
			}
			found := false
			ast.Inspect(fd.Body, func(n ast.Node) bool {
				if isAdminCall(n) {
					found = true
				}
				return !found
			})
			if found {
				sites[fd.Name.Name] = true
			}
		}
	}
	return sites
}

// answer is what a caller got from one admin call site, reduced to what
// must not differ between a non-admin and an admin in player mode: the
// status, the error text, and the shape of a success body (which
// fields came back non-empty, and every boolean).
type answer struct {
	Status int
	Error  string
	Shape  string
}

func (a answer) String() string { return fmt.Sprintf("%d %q {%s}", a.Status, a.Error, a.Shape) }

func shapeOf(v any, prefix string, out map[string]bool) {
	switch x := v.(type) {
	case map[string]any:
		for k, vv := range x {
			shapeOf(vv, prefix+"."+k, out)
		}
	case []any:
		for _, vv := range x {
			shapeOf(vv, prefix+"[]", out)
		}
	case string:
		if x != "" {
			out[prefix] = true
		}
	case bool:
		out[prefix+"="+strconv.FormatBool(x)] = true
	}
}

func answerOf(resp *http.Response) answer {
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	a := answer{Status: resp.StatusCode}
	var body any
	if json.Unmarshal(raw, &body) == nil {
		if m, ok := body.(map[string]any); ok {
			if e, ok := m["error"].(string); ok {
				a.Error = e
			}
		}
		set := map[string]bool{}
		shapeOf(body, "", set)
		keys := make([]string, 0, len(set))
		for k := range set {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		a.Shape = strings.Join(keys, " ")
	}
	return a
}

// sameAnswerCaller is who asks. The three differ only in the Discord
// ID's place on the list and in the mode.
type sameAnswerCaller struct {
	name      string
	discordID string
	modeOn    bool
}

var (
	callerNonAdmin = sameAnswerCaller{"a non-admin signed-in person", unlistedDiscordID, false}
	callerPlayer   = sameAnswerCaller{"an allowlisted person in player mode", listedDiscordID, false}
	callerAdmin    = sameAnswerCaller{"an allowlisted person in admin mode", listedDiscordID, true}
)

// Where the caller's session stands relative to the table the probe
// asks about.
const (
	sessionSignedIn  = "signed in, no seat"
	sessionElsewhere = "seated at another table"
	sessionHere      = "seated at this table, not its host"
)

// probeFixture is one fresh stack per (probe, caller, session): a
// table someone else created with two guests, Alice (its host) and
// Bob; a second table; and the caller's session.
type probeFixture struct {
	s          *adminStack
	table      GameMeta
	alice, bob uuid.UUID
	home       GameMeta
	user       uuid.UUID
	seat       uuid.UUID // the caller's seat, if any
	tok        string
	principal  auth.Principal
}

func (fx *probeFixture) call(t *testing.T, method, path string, body any) answer {
	t.Helper()
	return answerOf(do(t, fx.s.srv, method, path, fx.tok, body))
}

func (fx *probeFixture) tablePath(rest string) string {
	return "/games/" + fx.table.ID.String() + rest
}

func newProbeFixture(t *testing.T, idx *cards.Index, caller sameAnswerCaller, session string) *probeFixture {
	t.Helper()
	host := newFakeBotHost()
	user := uuid.New()
	// A replay directory, so bugReplaySource has a replay to offer.
	s := newAdminStackIn(t, t.TempDir(), idx, func(c *Config) {
		c.Lobby.SetBotHost(host)
		c.Bots = host
		c.BotDecks = fakeDeckSource{}
		// POST /deck-requests: the caller's account has its Discord
		// identity, has already asked three times today, and no deck
		// link reaches the network.
		c.Users = probeUsers{subjects: map[uuid.UUID]string{user: caller.discordID}}
		c.DeckRequests = fullDeckRequests{}
		c.DeckRequestFiler = &fakeFiler{issues: map[int]*fakeIssue{}}
		c.FetchDeck = func(context.Context, string) (string, []deck.Entry, error) {
			return "", nil, fmt.Errorf("%w: moxfield", deck.ErrDeckNotFound)
		}
		// The admin views (ADR 0124): a store with no rows, so admin
		// mode reaches each handler's answer.
		c.AdminViews = emptyAdminViews{}
	}, listedDiscordID)
	fx := &probeFixture{s: s, user: user}
	var err error
	if fx.table, err = s.lobby.Create("their table"); err != nil {
		t.Fatal(err)
	}
	if _, fx.alice, err = s.lobby.Join(fx.table.ID, fx.table.InviteToken, "Alice"); err != nil {
		t.Fatal(err)
	}
	if _, fx.bob, err = s.lobby.Join(fx.table.ID, fx.table.InviteToken, "Bob"); err != nil {
		t.Fatal(err)
	}
	if fx.home, err = s.lobby.Create("another table"); err != nil {
		t.Fatal(err)
	}
	fx.principal = auth.Principal{UserID: fx.user, DiscordID: caller.discordID, DiscordGlobalName: "Caller", Name: "Caller"}
	switch session {
	case sessionSignedIn:
		fx.principal.Role = auth.RoleIdentified
	case sessionElsewhere, sessionHere:
		at := fx.home
		if session == sessionHere {
			at = fx.table
		}
		if _, fx.seat, err = s.lobby.Join(at.ID, at.InviteToken, "Caller"); err != nil {
			t.Fatal(err)
		}
		fx.principal.Role = auth.RolePlayer
		fx.principal.GameID = at.ID
		fx.principal.PlayerID = fx.seat
	}
	fx.tok = s.issue(t, fx.principal)
	fx.principal, err = s.auth.Validate(context.Background(), fx.tok)
	if err != nil {
		t.Fatal(err)
	}
	if caller.modeOn {
		s.adminModeOn(t, fx.user)
	}
	return fx
}

// probeUsers is a user store that knows only the probe caller's Discord
// identity, which is all POST /deck-requests asks of it. Everything
// else answers as a deployment with no database does.
type probeUsers struct {
	users.NoStore
	subjects map[uuid.UUID]string
}

func (u probeUsers) DiscordSubject(_ context.Context, id uuid.UUID) (string, error) {
	if s, ok := u.subjects[id]; ok {
		return s, nil
	}
	return "", users.ErrNotFound
}

// fullDeckRequests is a deck-request store in which every requester has
// already used the day's three asks.
type fullDeckRequests struct{ deckrequests.NoStore }

func (fullDeckRequests) AsksSince(_ context.Context, _ string, since time.Time) ([]time.Time, error) {
	at := since.Add(time.Hour)
	return []time.Time{at, at, at}, nil
}

// label names an id from the fixture, so bindings compare across stacks.
func (fx *probeFixture) label(id uuid.UUID) string {
	switch id {
	case uuid.Nil:
		return "none"
	case fx.table.ID:
		return "the table"
	case fx.home.ID:
		return "another table"
	case fx.alice:
		return "alice"
	case fx.bob:
		return "bob"
	case fx.seat:
		return "own seat"
	case fx.user:
		return "own user"
	}
	return "other"
}

func (fx *probeFixture) upgrade(query string) answer {
	b, err := fx.s.wsAuth.AuthorizeUpgrade(upgradeRequest(fx.tok, query))
	if err != nil {
		// The status lives in an unexported type; the message names it.
		return answer{Error: err.Error()}
	}
	return answer{Status: http.StatusSwitchingProtocols, Shape: fmt.Sprintf("game=%s player=%s user=%s read_only=%v admin=%v",
		fx.label(b.GameID), fx.label(b.PlayerID), fx.label(b.UserID), b.ReadOnly, b.Admin)}
}

type sameAnswerProbe struct {
	name string
	// sites are the admin call sites (adminCallSites' names) this probe
	// reaches.
	sites []string
	// differs: admin mode answers differently for at least one session,
	// which proves the probe reaches the admin decision at all.
	differs bool
	ask     func(t *testing.T, fx *probeFixture) answer
}

func sameAnswerProbes() []sameAnswerProbe {
	httpProbe := func(method, path string, body func(fx *probeFixture) any) func(t *testing.T, fx *probeFixture) answer {
		return func(t *testing.T, fx *probeFixture) answer {
			var b any
			if body != nil {
				b = body(fx)
			}
			return fx.call(t, method, strings.ReplaceAll(path, "{id}", fx.table.ID.String()), b)
		}
	}
	return []sameAnswerProbe{
		// Routes behind requireAdmin.
		{"DELETE /games/{id}", []string{"route:DELETE /games/{id}"}, true, httpProbe("DELETE", "/games/{id}", nil)},
		{"POST /games/{id}/archive", []string{"route:POST /games/{id}/archive"}, true, httpProbe("POST", "/games/{id}/archive", nil)},
		{"DELETE /games/{id}/archive", []string{"route:DELETE /games/{id}/archive"}, true, httpProbe("DELETE", "/games/{id}/archive", nil)},
		{"POST /games/{id}/seats/{player}/reclaim", []string{"route:POST /games/{id}/seats/{player}/reclaim"}, true, func(t *testing.T, fx *probeFixture) answer {
			return fx.call(t, "POST", fx.tablePath("/seats/"+fx.alice.String()+"/reclaim"), nil)
		}},
		{"GET /games/{id}/creator", []string{"route:GET /games/{id}/creator"}, true, httpProbe("GET", "/games/{id}/creator?discord_id="+listedDiscordID, nil)},
		{"GET /bugreport/{id}/replay", []string{"route:GET /bugreport/{id}/replay"}, true, httpProbe("GET", "/bugreport/abc/replay", nil)},
		{"GET /bugreport/{id}/gamelog", []string{"route:GET /bugreport/{id}/gamelog"}, true, httpProbe("GET", "/bugreport/abc/gamelog", nil)},
		{"POST /admin/users/{id}/revoke-sessions", []string{"route:POST /admin/users/{id}/revoke-sessions"}, true, func(t *testing.T, fx *probeFixture) answer {
			return fx.call(t, "POST", "/admin/users/"+uuid.New().String()+"/revoke-sessions", nil)
		}},
		{"GET /games/{id}/bot/stats", []string{"BotStatsHandler"}, true, httpProbe("GET", "/games/{id}/bot/stats", nil)},
		// The admin views (ADR 0124 §8).
		{"GET /admin/users", []string{"route:GET /admin/users"}, true, httpProbe("GET", "/admin/users?played=7d", nil)},
		{"GET /admin/users/{id}", []string{"route:GET /admin/users/{id}"}, true, func(t *testing.T, fx *probeFixture) answer {
			return fx.call(t, "GET", "/admin/users/"+fx.user.String(), nil)
		}},
		{"GET /admin/games", []string{"route:GET /admin/games"}, true, httpProbe("GET", "/admin/games?practice=include", nil)},
		{"GET /admin/games/{id}", []string{"route:GET /admin/games/{id}"}, true, func(t *testing.T, fx *probeFixture) answer {
			return fx.call(t, "GET", "/admin/games/"+fx.table.ID.String(), nil)
		}},

		// Handler checks through c.isAdmin.
		{"POST /games/{id}/spawn", []string{"requireTableManager"}, true, httpProbe("POST", "/games/{id}/spawn", func(*probeFixture) any {
			return map[string]any{"name": "Plains", "zone": "hand", "count": 1}
		})},
		{"GET /games/{id}/spawn/tokens", []string{"requireTableManager"}, true, httpProbe("GET", "/games/{id}/spawn/tokens", nil)},
		{"GET /games/{id}/spawn/cards", []string{"requireTableManager"}, true, httpProbe("GET", "/games/{id}/spawn/cards?q=plains", nil)},
		{"POST /games", []string{"createGameWith"}, false, func(t *testing.T, fx *probeFixture) answer {
			return fx.call(t, "POST", "/games", createGameRequest{Name: "new table"})
		}},
		{"POST /games/{id}/host", []string{"transferHost"}, true, httpProbe("POST", "/games/{id}/host", func(fx *probeFixture) any {
			return transferHostRequest{PlayerID: fx.bob}
		})},
		{"PATCH /games/{id}/settings", []string{"updateTableSettings"}, true, httpProbe("PATCH", "/games/{id}/settings", func(*probeFixture) any {
			return map[string]any{"undo_limit": 3}
		})},
		{"GET /games", []string{"listGames"}, false, httpProbe("GET", "/games", nil)},
		{"GET /games?archived=1", []string{"listGames"}, false, httpProbe("GET", "/games?archived=1", nil)},
		{"GET /games/{id}", []string{"getGame"}, true, httpProbe("GET", "/games/{id}", nil)},
		{"POST /games/{id}/invites/rotate", []string{"rotateInvite"}, true, httpProbe("POST", "/games/{id}/invites/rotate", func(*probeFixture) any {
			return rotateInviteRequest{Kind: "player"}
		})},
		{"GET /games/{id}/replay", []string{"downloadReplay"}, true, httpProbe("GET", "/games/{id}/replay", nil)},
		{"POST /games/{id}/start", []string{"startGame"}, true, httpProbe("POST", "/games/{id}/start", nil)},
		{"POST /games/{id}/decks for Alice's seat", []string{"uploadDeck"}, true, httpProbe("POST", "/games/{id}/decks", func(fx *probeFixture) any {
			return uploadDeckRequest{Format: "text", Source: "Commander:\n1 Test Commander\nMainboard:\n99 Plains\n", PlayerID: fx.alice}
		})},
		{"POST /games/{id}/seats/bot", []string{"botSeatAuthorised"}, true, httpProbe("POST", "/games/{id}/seats/bot", func(*probeFixture) any {
			return addBotRequest{Tier: "random", Deck: "test-mono-white"}
		})},
		{"DELETE /games/{id}/seats/bot/{player}", []string{"botSeatAuthorised"}, true, func(t *testing.T, fx *probeFixture) answer {
			return fx.call(t, "DELETE", fx.tablePath("/seats/bot/"+fx.bob.String()), nil)
		}},
		// /me: admin reads false in player mode. admin_allowed stays
		// true for the allowlisted person, on purpose (§2 item 9), so
		// the comparison is of `admin` alone.
		{"GET /me", []string{"adminModeOf"}, true, func(t *testing.T, fx *probeFixture) answer {
			resp := doGet(t, fx.s.srv, "/me", fx.tok)
			defer resp.Body.Close()
			var me meResponse
			_ = json.NewDecoder(resp.Body).Decode(&me)
			return answer{Status: resp.StatusCode, Shape: fmt.Sprintf("admin=%v", me.Admin)}
		}},
		{"POST /games/{id}/setup", []string{"canApplySetup", "applySetupRoute"}, true, httpProbe("POST", "/games/{id}/setup", func(*probeFixture) any {
			return applySetupRequest{From: setupFromLast}
		})},
		{"POST /games/{id}/invites/dm", []string{"inviteDM"}, true, httpProbe("POST", "/games/{id}/invites/dm", func(*probeFixture) any {
			return dmInviteRequest{UserID: uuid.New().String()}
		})},
		// #2052: an admin is not rate-limited on deck requests. The
		// caller has used the day's three asks, so anyone else is a 429;
		// admin mode gets past it to the (failing) deck fetch.
		{"POST /deck-requests", []string{"deckRequestExempt"}, true, httpProbe("POST", "/deck-requests", func(*probeFixture) any {
			return map[string]string{"url": "https://moxfield.com/decks/probe01"}
		})},
		// The bug report's attachments, asked directly: a report from
		// one table must not pull another table's log or replay.
		{"bugGameLog", []string{"bugGameLog"}, true, func(t *testing.T, fx *probeFixture) answer {
			seedGameLog(t, fx.s.lobby, fx.table.ID)
			out := bugGameLog(fx.s.cfg, fx.principal, &bugReportContext{GameID: fx.table.ID.String()})
			return answer{Shape: fmt.Sprintf("log=%v", len(out) > 0)}
		}},
		{"bugReplaySource", []string{"bugReplaySource"}, true, func(t *testing.T, fx *probeFixture) answer {
			out := bugReplaySource(fx.s.cfg, fx.principal, &bugReportContext{GameID: fx.table.ID.String()})
			return answer{Shape: fmt.Sprintf("replay=%v", out != "")}
		}},
		// The WebSocket upgrade.
		{"WS ?game=<the table>", []string{"AuthorizeUpgrade"}, true, func(t *testing.T, fx *probeFixture) answer {
			return fx.upgrade("game=" + fx.table.ID.String())
		}},
		{"WS ?game=<the table>&player=<alice>", []string{"AuthorizeUpgrade"}, true, func(t *testing.T, fx *probeFixture) answer {
			return fx.upgrade("game=" + fx.table.ID.String() + "&player=" + fx.alice.String())
		}},
		{"WS ?game=<another table>", []string{"AuthorizeUpgrade"}, true, func(t *testing.T, fx *probeFixture) answer {
			return fx.upgrade("game=" + fx.home.ID.String())
		}},
		{"WS, no query", []string{"AuthorizeUpgrade"}, true, func(t *testing.T, fx *probeFixture) answer {
			return fx.upgrade("")
		}},
	}
}

// TestPlayerModeAnswersExactlyAsANonAdmin is ADR 0112 §2's table: every
// admin call site, asked by a non-admin signed-in person, by an
// allowlisted person in player mode and by one in admin mode, from
// three places (no seat, a seat at another table, a seat at this
// table). The first two answers must be identical everywhere.
func TestPlayerModeAnswersExactlyAsANonAdmin(t *testing.T) {
	t.Setenv("CMDCTRL_DEV_RELAX_RATE_LIMITS", "1")
	idx := buildMinimalDeckIndex(t)
	for _, probe := range sameAnswerProbes() {
		t.Run(probe.name, func(t *testing.T) {
			differed := false
			var seen []string
			for _, session := range []string{sessionSignedIn, sessionElsewhere, sessionHere} {
				nonAdmin := probe.ask(t, newProbeFixture(t, idx, callerNonAdmin, session))
				player := probe.ask(t, newProbeFixture(t, idx, callerPlayer, session))
				admin := probe.ask(t, newProbeFixture(t, idx, callerAdmin, session))
				if nonAdmin != player {
					t.Errorf("%s:\n  %s got %s\n  %s got %s\n  player mode must answer exactly as a non-admin",
						session, callerNonAdmin.name, nonAdmin, callerPlayer.name, player)
				}
				if admin != player {
					differed = true
				}
				seen = append(seen, fmt.Sprintf("%s: player %s, admin %s", session, player, admin))
			}
			if probe.differs && !differed {
				t.Errorf("admin mode answered exactly as player mode from every session, so this probe never reached the admin decision:\n  %s", strings.Join(seen, "\n  "))
			}
		})
	}
}

// TestEveryAdminCallSiteIsInTheSameAnswerTable: a new admin decision
// fails here until a probe in the table above asks it.
func TestEveryAdminCallSiteIsInTheSameAnswerTable(t *testing.T) {
	want := adminCallSites(t)
	covered := map[string]bool{}
	for _, p := range sameAnswerProbes() {
		for _, s := range p.sites {
			covered[s] = true
		}
	}
	for site := range want {
		if !covered[site] {
			t.Errorf("admin call site %q is not asked by TestPlayerModeAnswersExactlyAsANonAdmin; add a probe for it", site)
		}
	}
	for site := range covered {
		if !want[site] {
			t.Errorf("the same-answer table names %q, which is no longer an admin call site", site)
		}
	}
	if len(want) < 25 {
		t.Errorf("found only %d admin call sites; the census is not reading the package", len(want))
	}
}

// parseNonTestGoFiles parses every non-test Go file in dir. It replaces
// go/parser.ParseDir, deprecated since Go 1.25; like the call it
// replaces, it ignores build tags.
func parseNonTestGoFiles(fset *token.FileSet, dir string) ([]*ast.File, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var files []*ast.File
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, filepath.Join(dir, name), nil, 0)
		if err != nil {
			return nil, err
		}
		files = append(files, f)
	}
	return files, nil
}
