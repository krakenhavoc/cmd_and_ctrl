package lobby

// archive_closed_test.go: an archived table is closed, not just hidden
// (#4564, ADR 0211 §4 "Closed means closed"). Every way into a table
// refuses it with ErrGameArchived, mapped to 409 game_closed, and an
// admin's socket is still admitted so the operator can look at it.
// Unarchiving opens every path again.

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/google/uuid"
)

// A lobby table that was archived refuses a seat, a spectator, a deck,
// a bot and a start; unarchiving it lets each one through again.
func TestArchivedLobbyTableRefusesEveryEntryPath(t *testing.T) {
	l := newTestLobby(t)
	meta, _ := l.Create("Closed table")
	_, alice, err := l.Join(meta.ID, meta.InviteToken, "Alice")
	if err != nil {
		t.Fatalf("Join Alice: %v", err)
	}
	if _, err := l.SetArchived(meta.ID, true); err != nil {
		t.Fatalf("archive: %v", err)
	}

	refused := map[string]error{}
	_, _, refused["Join"] = l.Join(meta.ID, meta.InviteToken, "Bob")
	_, _, refused["JoinAgent"] = l.JoinAgent(meta.ID, meta.InviteToken, "Claude", AgentDecl{Client: "test"})
	_, refused["Spectate"] = l.Spectate(meta.ID, meta.SpectatorInvite)
	_, refused["SetDeck"] = l.SetDeck(meta.ID, alice, "Alice's", botDeck(20))
	refused["SetSeatDeckID"] = l.SetSeatDeckID(meta.ID, alice, uuid.NewString(), 3)
	_, _, refused["AddBot"] = l.AddBot(meta.ID, "Bot 1", "random", "", "Mono Red", botDeck(20))
	_, refused["Start"] = l.Start(meta.ID)
	for name, err := range refused {
		if !errors.Is(err, ErrGameArchived) {
			t.Errorf("%s on an archived table: %v, want ErrGameArchived", name, err)
		}
	}
	got, _ := l.Get(meta.ID)
	if len(got.Players) != 1 {
		t.Errorf("archived table gained seats: %d, want 1", len(got.Players))
	}

	if _, err := l.SetArchived(meta.ID, false); err != nil {
		t.Fatalf("unarchive: %v", err)
	}
	if _, err := l.Spectate(meta.ID, meta.SpectatorInvite); err != nil {
		t.Errorf("Spectate after unarchive: %v", err)
	}
	if _, err := l.SetDeck(meta.ID, alice, "Alice's", botDeck(20)); err != nil {
		t.Errorf("SetDeck after unarchive: %v", err)
	}
	if err := l.SetSeatDeckID(meta.ID, alice, "", 0); err != nil {
		t.Errorf("SetSeatDeckID after unarchive: %v", err)
	}
	if _, _, err := l.AddBot(meta.ID, "Bot 1", "random", "", "Mono Red", botDeck(20)); err != nil {
		t.Errorf("AddBot after unarchive: %v", err)
	}
	if _, _, err := l.Join(meta.ID, meta.InviteToken, "Bob"); err != nil {
		t.Errorf("Join after unarchive: %v", err)
	}
}

// Start on a running game is an idempotent no-op that returns the
// meta; on an archived one it must refuse instead, so a client cannot
// read a closed game as live.
func TestArchivedRunningGameRefusesStart(t *testing.T) {
	l := newTestLobby(t)
	meta, _ := l.Create("Running then closed")
	_, alice, err := l.Join(meta.ID, meta.InviteToken, "Alice")
	if err != nil {
		t.Fatal(err)
	}
	_, bob, err := l.Join(meta.ID, meta.InviteToken, "Bob")
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range []uuid.UUID{alice, bob} {
		if _, err := l.SetDeck(meta.ID, p, "deck", botDeck(20)); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := l.Start(meta.ID); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if _, err := l.SetArchived(meta.ID, true); err != nil {
		t.Fatal(err)
	}
	if _, err := l.Start(meta.ID); !errors.Is(err, ErrGameArchived) {
		t.Errorf("Start on an archived running game: %v, want ErrGameArchived", err)
	}
}

// The refusal is the invite's to see: a wrong invite on an archived
// table still reads as a wrong invite, so the 409 tells nobody without
// a credential that the table exists.
func TestArchivedTableChecksTheInviteFirst(t *testing.T) {
	l := newTestLobby(t)
	meta, _ := l.Create("Closed table")
	if _, err := l.SetArchived(meta.ID, true); err != nil {
		t.Fatal(err)
	}
	if _, _, err := l.Join(meta.ID, "wrong", "Bob"); !errors.Is(err, ErrInvalidInvite) {
		t.Errorf("Join with a wrong invite: %v, want ErrInvalidInvite", err)
	}
	if _, err := l.Spectate(meta.ID, "wrong"); !errors.Is(err, ErrInvalidInvite) {
		t.Errorf("Spectate with a wrong invite: %v, want ErrInvalidInvite", err)
	}
}

// Over HTTP the refusal is 409 with code "game_closed" (ADR 0211 §4).
func TestArchivedTableJoinAndSpectateAre409GameClosed(t *testing.T) {
	srv, l, _ := newTestHTTPStack(t)
	meta, _ := l.Create("Closed table")
	if _, err := l.SetArchived(meta.ID, true); err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		path string
		body any
	}{
		{"/games/" + meta.ID.String() + "/join", joinRequest{InviteToken: meta.InviteToken, Name: "Bob"}},
		{"/games/" + meta.ID.String() + "/spectate", spectateRequest{InviteToken: meta.SpectatorInvite, Name: "watcher"}},
	} {
		resp := postJSON(t, srv, tc.path, "", tc.body)
		var body struct {
			Error string `json:"error"`
			Code  string `json:"code"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&body)
		resp.Body.Close()
		if resp.StatusCode != http.StatusConflict || body.Code != CodeGameClosed {
			t.Errorf("POST %s on an archived table: %d %+v, want 409 code %q",
				tc.path, resp.StatusCode, body, CodeGameClosed)
		}
	}
}

// A seat and a spectator who still hold sessions for a table that was
// archived cannot reconnect to it; an admin still can, and unarchiving
// lets the seat back in.
func TestArchivedTableRefusesAWebSocketUpgradeButNotAnAdmin(t *testing.T) {
	srv, l, _ := newTestHTTPStack(t)
	meta, _ := l.Create("Closed table")

	resp := postJSON(t, srv, "/games/"+meta.ID.String()+"/join", "",
		joinRequest{InviteToken: meta.InviteToken, Name: "Alice"})
	var seat sessionResponse
	_ = json.NewDecoder(resp.Body).Decode(&seat)
	resp.Body.Close()
	resp = postJSON(t, srv, "/games/"+meta.ID.String()+"/spectate", "",
		spectateRequest{InviteToken: meta.SpectatorInvite, Name: "watcher"})
	var spec sessionResponse
	_ = json.NewDecoder(resp.Body).Decode(&spec)
	resp.Body.Close()
	resp = postJSON(t, srv, "/admin/login", "", adminLoginRequest{Token: "shared-admin-token"})
	var admin sessionResponse
	_ = json.NewDecoder(resp.Body).Decode(&admin)
	resp.Body.Close()
	if seat.Token == "" || spec.Token == "" || admin.Token == "" {
		t.Fatalf("setup: seat=%q spectator=%q admin=%q", seat.Token, spec.Token, admin.Token)
	}

	game := "game=" + meta.ID.String()
	if got := dialStatus(t, srv, seat.Token, game); got != http.StatusSwitchingProtocols {
		t.Fatalf("seat before archive: %d, want 101", got)
	}
	if _, err := l.SetArchived(meta.ID, true); err != nil {
		t.Fatal(err)
	}

	if got := dialStatus(t, srv, seat.Token, game); got != http.StatusConflict {
		t.Errorf("seat on an archived table: %d, want 409", got)
	}
	if got := dialStatus(t, srv, spec.Token, game); got != http.StatusConflict {
		t.Errorf("spectator on an archived table: %d, want 409", got)
	}
	if got := dialStatus(t, srv, admin.Token, game); got != http.StatusSwitchingProtocols {
		t.Errorf("admin on an archived table: %d, want 101", got)
	}

	if _, err := l.SetArchived(meta.ID, false); err != nil {
		t.Fatal(err)
	}
	if got := dialStatus(t, srv, seat.Token, game); got != http.StatusSwitchingProtocols {
		t.Errorf("seat after unarchive: %d, want 101", got)
	}
}
