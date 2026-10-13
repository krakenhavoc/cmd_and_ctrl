package lobby

// reports_test.go: "report player" (#2984).

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// reportStack is the HTTP wiring WITH the rate limiters on (unlike
// seatRemoveStack, which relaxes them for tests that make many calls).
type reportStack struct {
	srv *httptest.Server
	l   *Lobby
	a   auth.Authenticator
}

func newReportStack(t *testing.T, store Store) *reportStack {
	t.Helper()
	t.Setenv("CMDCTRL_DEV_RELAX_RATE_LIMITS", "")
	mgr := ws.NewRoomManager(quietLogger(), "")
	var l *Lobby
	if store != nil {
		l = NewLobbyWithStore(mgr, store)
	} else {
		l = NewLobby(mgr)
	}
	a := newHMAC(t)
	srv := httptest.NewServer(Handler(Config{Lobby: l, Auth: a, AdminToken: "shared-admin-token"}))
	t.Cleanup(srv.Close)
	return &reportStack{srv: srv, l: l, a: a}
}

func (s *reportStack) report(t *testing.T, gameID uuid.UUID, token string, target uuid.UUID, reason string) *http.Response {
	t.Helper()
	return postJSON(t, s.srv, "/games/"+gameID.String()+"/reports", token,
		map[string]string{"player_id": target.String(), "reason": reason})
}

func wantCode(t *testing.T, resp *http.Response, want int) {
	t.Helper()
	defer resp.Body.Close()
	if resp.StatusCode != want {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("status %d, want %d: %s", resp.StatusCode, want, body)
	}
}

// reportTable seats Alice, Bob, Carol and returns their sessions.
func (s *reportStack) reportTable(t *testing.T) (GameMeta, sessionResponse, sessionResponse, sessionResponse) {
	t.Helper()
	meta, err := s.l.Create("Report table")
	if err != nil {
		t.Fatal(err)
	}
	return meta, joinAs(t, s.srv, meta, "Alice"), joinAs(t, s.srv, meta, "Bob"), joinAs(t, s.srv, meta, "Carol")
}

func (s *reportStack) adminList(t *testing.T) adminReportsResponse {
	t.Helper()
	resp := doGet(t, s.srv, "/admin/reports", adminSession(t, s.a))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("admin list: %d %s", resp.StatusCode, body)
	}
	var out adminReportsResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	return out
}

func TestReportIsStoredWithTheRecentChatAndListedForTheAdmin(t *testing.T) {
	s := newReportStack(t, nil)
	meta, alice, bob, _ := s.reportTable(t)
	room := s.l.RoomOf(meta.ID)
	for i := 0; i < ws.RecentChatLimit+5; i++ {
		room.RecordChat(protocol.ChatPayload{AuthorName: "Bob", Text: fmt.Sprintf("line %d", i), Timestamp: "2026-10-13T00:00:00Z"})
	}

	resp := s.report(t, meta.ID, alice.Token, bob.PlayerID, "  slurs in chat  ")
	wantCode(t, resp, http.StatusCreated)

	got := s.adminList(t)
	if len(got.Reports) != 1 {
		t.Fatalf("reports = %d, want 1", len(got.Reports))
	}
	r := got.Reports[0]
	if r.GameID != meta.ID.String() || r.GameName != "Report table" {
		t.Errorf("table = %q %q", r.GameID, r.GameName)
	}
	if r.Reporter.Name != "Alice" || r.Reporter.PlayerID != alice.PlayerID.String() ||
		r.Reported.Name != "Bob" || r.Reported.PlayerID != bob.PlayerID.String() {
		t.Errorf("parties = %+v / %+v", r.Reporter, r.Reported)
	}
	if r.Reason != "slurs in chat" {
		t.Errorf("reason = %q, want it trimmed", r.Reason)
	}
	if len(r.Excerpt) != reportExcerptLines {
		t.Fatalf("excerpt has %d lines, want the last %d", len(r.Excerpt), reportExcerptLines)
	}
	if last := r.Excerpt[len(r.Excerpt)-1]; last.Text != fmt.Sprintf("line %d", ws.RecentChatLimit+4) || last.Author != "Bob" {
		t.Errorf("last excerpt line = %+v", last)
	}
	if got.RetentionDays != 90 {
		t.Errorf("retention_days = %d", got.RetentionDays)
	}
	if time.Since(time.UnixMilli(r.CreatedAt)) > time.Minute {
		t.Errorf("created_at = %d", r.CreatedAt)
	}
}

func TestReportListIsAdminOnly(t *testing.T) {
	s := newReportStack(t, nil)
	meta, alice, bob, _ := s.reportTable(t)
	wantCode(t, s.report(t, meta.ID, alice.Token, bob.PlayerID, "rude"), http.StatusCreated)

	spec := decodeSession(t, postJSON(t, s.srv, "/games/"+meta.ID.String()+"/spectate", "",
		map[string]string{"invite_token": meta.SpectatorInvite}), http.StatusOK)
	for name, tok := range map[string]string{
		"anonymous":      "",
		"the reporter":   alice.Token,
		"the reported":   bob.Token,
		"a spectator":    spec.Token,
		"a signed-in":    identifiedToken(t, s.a, uuid.New()),
		"the table host": alice.Token,
	} {
		want := http.StatusForbidden
		if tok == "" {
			want = http.StatusUnauthorized
		}
		t.Run(name, func(t *testing.T) {
			wantCode(t, doGet(t, s.srv, "/admin/reports", tok), want)
			wantCode(t, doJSON(t, s.srv, http.MethodDelete, "/admin/reports/"+uuid.NewString(), tok, nil), want)
		})
	}
	if n := len(s.adminList(t).Reports); n != 1 {
		t.Errorf("reports = %d after refused deletes", n)
	}
}

func TestReportRefusals(t *testing.T) {
	s := newReportStack(t, nil)
	meta, alice, bob, carol := s.reportTable(t)
	other, _ := s.l.Create("Elsewhere")
	zed := joinAs(t, s.srv, other, "Zed")
	spec := decodeSession(t, postJSON(t, s.srv, "/games/"+meta.ID.String()+"/spectate", "",
		map[string]string{"invite_token": meta.SpectatorInvite}), http.StatusOK)

	wantCode(t, s.report(t, meta.ID, "", bob.PlayerID, "x"), http.StatusUnauthorized)
	wantCode(t, s.report(t, meta.ID, spec.Token, bob.PlayerID, "x"), http.StatusForbidden)           // a spectator has no seat
	wantCode(t, s.report(t, meta.ID, adminSession(t, s.a), bob.PlayerID, "x"), http.StatusForbidden) // nor does the admin token
	wantCode(t, s.report(t, meta.ID, zed.Token, bob.PlayerID, "x"), http.StatusForbidden)            // a seat at ANOTHER table
	wantCode(t, s.report(t, meta.ID, alice.Token, alice.PlayerID, "x"), http.StatusUnprocessableEntity)
	wantCode(t, s.report(t, meta.ID, alice.Token, uuid.New(), "x"), http.StatusUnprocessableEntity)
	wantCode(t, s.report(t, meta.ID, alice.Token, bob.PlayerID, "   "), http.StatusBadRequest)
	// (A refused attempt still spends the reporter's allowance, so Carol
	// takes this one: Alice is out of tokens by now.)
	resp := postJSON(t, s.srv, "/games/"+meta.ID.String()+"/reports", carol.Token, map[string]string{"player_id": "not-a-uuid", "reason": "x"})
	wantCode(t, resp, http.StatusBadRequest)
	if n := len(s.adminList(t).Reports); n != 0 {
		t.Errorf("a refused report was stored: %d", n)
	}
}

func TestReportsAreRateLimitedPerReporter(t *testing.T) {
	s := newReportStack(t, nil)
	meta, alice, bob, carol := s.reportTable(t)

	// Burst of three, then 429 for Alice.
	for i := 0; i < 3; i++ {
		wantCode(t, s.report(t, meta.ID, alice.Token, bob.PlayerID, "again"), http.StatusCreated)
	}
	resp := s.report(t, meta.ID, alice.Token, bob.PlayerID, "again")
	wantCode(t, resp, http.StatusTooManyRequests)
	if got := len(s.adminList(t).Reports); got != 3 {
		t.Errorf("reports = %d, want 3: the limited one must not be stored", got)
	}
	// Another reporter at the same table still has their own allowance.
	wantCode(t, s.report(t, meta.ID, carol.Token, bob.PlayerID, "me too"), http.StatusCreated)
}

func TestAdminDismissesAReport(t *testing.T) {
	s := newReportStack(t, nil)
	meta, alice, bob, _ := s.reportTable(t)
	wantCode(t, s.report(t, meta.ID, alice.Token, bob.PlayerID, "rude"), http.StatusCreated)
	id := s.adminList(t).Reports[0].ID
	admin := adminSession(t, s.a)

	wantCode(t, doJSON(t, s.srv, http.MethodDelete, "/admin/reports/"+id, admin, nil), http.StatusNoContent)
	wantCode(t, doJSON(t, s.srv, http.MethodDelete, "/admin/reports/"+id, admin, nil), http.StatusNotFound)
	wantCode(t, doJSON(t, s.srv, http.MethodDelete, "/admin/reports/nope", admin, nil), http.StatusBadRequest)
	if n := len(s.adminList(t).Reports); n != 0 {
		t.Errorf("reports = %d after dismissing the only one", n)
	}
}

func TestReportsSurviveARestartAndOutliveTheirTable(t *testing.T) {
	dir := t.TempDir()
	st := NewSQLStore(openTestDB(t, dir))
	s := newReportStack(t, st)
	meta, alice, bob, _ := s.reportTable(t)
	s.l.RoomOf(meta.ID).RecordChat(protocol.ChatPayload{AuthorName: "Bob", Text: "hello", Timestamp: "t"})
	wantCode(t, s.report(t, meta.ID, alice.Token, bob.PlayerID, "rude"), http.StatusCreated)

	// A second process over the same database sees it.
	l2 := NewLobbyWithStore(ws.NewRoomManager(quietLogger(), dir), NewSQLStore(openTestDB(t, dir)))
	got, err := l2.Reports()
	if err != nil || len(got) != 1 || got[0].ReportedName != "Bob" || len(got[0].Excerpt) != 1 || got[0].Excerpt[0].Text != "hello" {
		t.Fatalf("after a restart: %+v, %v", got, err)
	}

	// Deleting the table does not delete what was said about it.
	if err := s.l.Delete(meta.ID); err != nil {
		t.Fatal(err)
	}
	if got, _ := l2.Reports(); len(got) != 1 {
		t.Errorf("a report went with its table: %d", len(got))
	}
}

func TestReportsExpireAndAreCapped(t *testing.T) {
	for name, store := range map[string]func(t *testing.T) Store{
		"memory": func(*testing.T) Store { return NewMemoryStore() },
		"sql":    func(t *testing.T) Store { return NewSQLStore(openTestDB(t, t.TempDir())) },
	} {
		t.Run(name, func(t *testing.T) {
			st := store(t)
			l := NewLobbyWithStore(ws.NewRoomManager(quietLogger(), ""), st)
			now := time.Now().UTC().Truncate(time.Millisecond)
			add := func(age time.Duration) uuid.UUID {
				id := uuid.New()
				if err := st.AddReport(t.Context(), ReportRecord{
					ID: id, CreatedAt: now.Add(-age), GameID: uuid.New(), GameName: "g",
					ReporterPlayerID: uuid.New(), ReporterName: "a", ReportedPlayerID: uuid.New(), ReportedName: "b", Reason: "r",
				}); err != nil {
					t.Fatal(err)
				}
				return id
			}
			fresh := add(time.Hour)
			add(ReportRetention + time.Hour)
			add(ReportRetention * 2)

			n, err := l.PruneReports(now)
			if err != nil || n != 2 {
				t.Fatalf("pruned %d, %v; want the two expired ones", n, err)
			}
			got, _ := l.Reports()
			if len(got) != 1 || got[0].ID != fresh {
				t.Fatalf("after expiry: %+v", got)
			}

			// The cap drops the oldest first.
			var newest uuid.UUID
			for i := 0; i < ReportMaxRows+7; i++ {
				newest = add(time.Duration(ReportMaxRows+8-i) * time.Minute)
			}
			if _, err := l.PruneReports(now); err != nil {
				t.Fatal(err)
			}
			got, _ = l.Reports()
			if len(got) != ReportMaxRows {
				t.Fatalf("reports = %d, want the cap %d", len(got), ReportMaxRows)
			}
			if got[0].ID != fresh && got[0].ID != newest {
				t.Errorf("newest first: got %v", got[0].ID)
			}
		})
	}
}
