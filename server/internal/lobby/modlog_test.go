package lobby

// Tests for modlog.go (ADR 0156's 2026-10-16 amendment, #3325): what the
// mod log posts for a player report and for each admin action, that a
// host's own moves are not logged, and that nothing it posts names a
// reporter, an account id or a Discord id.

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/discord"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// fakeModLog is the Discord half: it records every entry and answers
// from canned errors, so nothing here touches a network.
type fakeModLog struct {
	mu      sync.Mutex
	entries []discord.ModLogEntry
	errs    []error // consumed one per Post call, nil past the end
	calls   int
}

func (f *fakeModLog) Post(_ context.Context, _ *http.Client, e discord.ModLogEntry) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	if len(f.errs) > 0 {
		err := f.errs[0]
		f.errs = f.errs[1:]
		if err != nil {
			return err
		}
	}
	f.entries = append(f.entries, e)
	return nil
}

func (f *fakeModLog) posted() []discord.ModLogEntry {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]discord.ModLogEntry(nil), f.entries...)
}

func newTestModLog(api modLogAPI) *ModLog {
	m := newModLog(api, "https://cmd.example/", nil, quietLogger())
	m.wait = func(time.Duration) {}
	return m
}

// field reads one field's value by name, "" when it is absent.
func field(e discord.ModLogEntry, name string) string {
	for _, f := range e.Fields {
		if f.Name == name {
			return f.Value
		}
	}
	return ""
}

// noIDs fails when an entry carries any of the ids, anywhere.
func noIDs(t *testing.T, e discord.ModLogEntry, ids ...string) {
	t.Helper()
	all := e.Title + " " + e.Actor + " " + e.Footer
	for _, f := range e.Fields {
		all += " " + f.Name + " " + f.Value
	}
	for _, id := range ids {
		if id != "" && strings.Contains(all, id) {
			t.Errorf("entry %q carries %q: %+v", e.Title, id, e)
		}
	}
}

// modLogStack is the seat-removal wiring with a mod log behind it.
type modLogStack struct {
	*seatRemoveStack
	fake *fakeModLog
	log  *ModLog
}

func newModLogStack(t *testing.T) *modLogStack {
	t.Helper()
	t.Setenv("CMDCTRL_DEV_RELAX_RATE_LIMITS", "1")
	log := quietLogger()
	mgr := ws.NewRoomManager(log, "")
	l := NewLobby(mgr)
	a := auth.WithSeatCheck(newHMAC(t), l)
	hub := ws.NewHub(log)
	hub.SetManager(mgr)
	hub.SetAuthorizer(&WSAuthorizer{Auth: a})
	l.SetStateBroadcaster(hub)
	fake := &fakeModLog{}
	ml := newTestModLog(fake)
	mux := http.NewServeMux()
	mux.Handle("/", Handler(Config{
		Lobby:        l,
		Auth:         a,
		AdminToken:   "shared-admin-token",
		SeatEvictor:  hub,
		SpectatorHub: hub,
		ModLog:       ml,
	}))
	mux.HandleFunc("GET /ws", hub.ServeWS)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return &modLogStack{seatRemoveStack: &seatRemoveStack{srv: srv, l: l, a: a, hub: hub}, fake: fake, log: ml}
}

func TestModLogPostsAPlayerReportWithoutTheReporter(t *testing.T) {
	s := newModLogStack(t)
	meta, err := s.l.Create("Friday pod")
	if err != nil {
		t.Fatal(err)
	}
	alice := joinAs(t, s.srv, meta, "Alice")
	bob := joinAs(t, s.srv, meta, "Bob")

	wantCode(t, postJSON(t, s.srv, "/games/"+meta.ID.String()+"/reports", alice.Token,
		map[string]string{"player_id": bob.PlayerID.String(), "reason": "  rude in chat  "}), http.StatusCreated)
	s.log.Wait()

	got := s.fake.posted()
	if len(got) != 1 {
		t.Fatalf("entries = %+v, want one", got)
	}
	e := got[0]
	if e.Title != "Player report" || e.Link != "https://cmd.example/#/admin/reports" {
		t.Errorf("entry = %+v", e)
	}
	if field(e, "Table") != "Friday pod" || field(e, "Reported") != "Bob" || field(e, "Reason") != "rude in chat" {
		t.Errorf("fields = %+v", e.Fields)
	}
	if e.At.IsZero() {
		t.Error("a report entry has no time")
	}
	// The reporter is in the admin view and nowhere here.
	noIDs(t, e, "Alice", alice.PlayerID.String(), bob.PlayerID.String(), meta.ID.String())
}

func TestModLogPostsAnAdminRemovingASeatButNotTheHost(t *testing.T) {
	s := newModLogStack(t)
	meta, alice, bob := s.table(t)

	// The host's own table: not a mod action.
	s.wantStatus(t, s.del(t, meta.ID, bob.PlayerID, alice.Token), http.StatusOK)
	s.log.Wait()
	if got := s.fake.posted(); len(got) != 0 {
		t.Fatalf("the host removing a seat was logged: %+v", got)
	}

	carol := joinAs(t, s.srv, meta, "Carol")
	admin := adminSession(t, s.a)
	resp := doJSON(t, s.srv, http.MethodDelete,
		"/games/"+meta.ID.String()+"/seats/"+carol.PlayerID.String()+"?block=true", admin, nil)
	s.wantStatus(t, resp, http.StatusOK)
	s.log.Wait()

	got := s.fake.posted()
	if len(got) != 1 {
		t.Fatalf("entries = %+v, want one", got)
	}
	e := got[0]
	if e.Title != "Seat removed" || field(e, "Table") != "FNM" || field(e, "Seat") != "Carol" ||
		field(e, "Blocked") != "yes, from this table" {
		t.Errorf("entry = %+v", e)
	}
	if e.Link != "https://cmd.example/#/admin/games/"+meta.ID.String() {
		t.Errorf("link = %q", e.Link)
	}
	if !strings.Contains(e.Actor, "admin token") {
		t.Errorf("actor = %q", e.Actor)
	}
	noIDs(t, e, carol.PlayerID.String())
}

func TestModLogPostsTablesEndedRestoredAndDeleted(t *testing.T) {
	s := newModLogStack(t)
	meta, err := s.l.Create("Old table")
	if err != nil {
		t.Fatal(err)
	}
	admin := adminSession(t, s.a)
	path := "/games/" + meta.ID.String()

	s.wantStatus(t, doJSON(t, s.srv, http.MethodPost, path+"/archive", admin, nil), http.StatusOK)
	s.wantStatus(t, doJSON(t, s.srv, http.MethodDelete, path+"/archive", admin, nil), http.StatusOK)
	s.wantStatus(t, doJSON(t, s.srv, http.MethodDelete, path, admin, nil), http.StatusNoContent)
	s.log.Wait()

	got := s.fake.posted()
	titles := map[string]discord.ModLogEntry{}
	for _, e := range got {
		titles[e.Title] = e
	}
	if len(got) != 3 || len(titles) != 3 {
		t.Fatalf("entries = %+v, want ended, restored and deleted", got)
	}
	for _, title := range []string{"Table ended", "Table restored", "Table deleted"} {
		e, ok := titles[title]
		if !ok {
			t.Errorf("no %q entry in %+v", title, got)
			continue
		}
		if field(e, "Table") != "Old table" || e.Actor == "" {
			t.Errorf("%s: %+v", title, e)
		}
	}
	if titles["Table deleted"].Link != "" {
		t.Errorf("a deleted table links to its gone admin page: %q", titles["Table deleted"].Link)
	}
}

func TestModLogPostsRevokedSessionsByDisplayName(t *testing.T) {
	fake := &fakeModLog{}
	ml := newTestModLog(fake)
	s := newRevocationStack(t, func(c *Config) { c.ModLog = ml })
	alice := signInAs(t, s, "4242")
	aliceID := mustValidate(t, s.auth, alice).UserID

	resp := post(t, s.srv, "/admin/users/"+aliceID.String()+"/revoke-sessions", adminToken(t, s.srv))
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("revoke: %d", resp.StatusCode)
	}
	ml.Wait()

	got := fake.posted()
	if len(got) != 1 {
		t.Fatalf("entries = %+v, want one", got)
	}
	e := got[0]
	if e.Title != "Sessions revoked" || field(e, "Account") != "U" || field(e, "Connections closed") != "1" {
		t.Errorf("entry = %+v", e)
	}
	if e.Link != "https://cmd.example/#/admin/accounts/"+aliceID.String() {
		t.Errorf("link = %q", e.Link)
	}
	// The link is the admin view's; the text names nobody by id.
	noIDs(t, e, aliceID.String(), "4242")
}

func TestModLogRetriesATransientFailureOnly(t *testing.T) {
	fake := &fakeModLog{errs: []error{discord.ErrVoiceUnavailable, discord.ErrVoiceUnavailable}}
	m := newTestModLog(fake)
	m.TableDeleted(auth.Principal{Role: auth.RoleAdmin}, "T")
	m.Wait()
	if fake.calls != 3 || len(fake.posted()) != 1 {
		t.Errorf("calls %d, posted %d; want 3 and 1", fake.calls, len(fake.posted()))
	}

	fake = &fakeModLog{errs: []error{discord.ErrModLogForbidden}}
	m = newTestModLog(fake)
	m.TableDeleted(auth.Principal{Role: auth.RoleAdmin}, "T")
	m.Wait()
	if fake.calls != 1 || len(fake.posted()) != 0 {
		t.Errorf("a 403 was retried: calls %d", fake.calls)
	}
}

func TestModLogNilIsOff(t *testing.T) {
	var m *ModLog
	// None of these may panic or block.
	m.Report(ReportRecord{})
	m.TableArchived(auth.Principal{}, uuid.New(), "T", true)
	m.TableDeleted(auth.Principal{}, "T")
	m.SeatRemoved(auth.Principal{}, uuid.New(), "T", "S", false, false)
	m.SessionsRevoked(auth.Principal{}, uuid.New(), "", 0)
	m.Wait()
}

func TestModLogActor(t *testing.T) {
	if got := modLogActor(auth.Principal{Role: auth.RoleIdentified, UserID: uuid.New(), Name: "Ada"}); got != "Ada (admin)" {
		t.Errorf("allowlisted admin = %q", got)
	}
	if got := modLogActor(auth.Principal{Role: auth.RoleAdmin}); !strings.Contains(got, "admin token") {
		t.Errorf("shared token = %q", got)
	}
}

func TestModLogWithNoOriginHasNoLinks(t *testing.T) {
	fake := &fakeModLog{}
	m := newModLog(fake, "", nil, quietLogger())
	m.Report(ReportRecord{GameName: "T", ReportedName: "B", Reason: "r"})
	m.Wait()
	if got := fake.posted(); len(got) != 1 || got[0].Link != "" {
		t.Errorf("entries = %+v", got)
	}
}
