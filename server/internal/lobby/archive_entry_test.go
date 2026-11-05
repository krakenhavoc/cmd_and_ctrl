package lobby

// archive_entry_test.go is #4564: an archived table is retired, so no
// way back into it may reopen it. Before the fix only invite-code
// redemption (FindByInvite) and the seat-reclaim paths checked; a
// player invite still seated someone, a spectator invite still watched,
// a seated player could still set a deck, add a bot or start, and a
// session minted before the archive still reconnected its WebSocket,
// all on a table that costs nothing against the load cap (capacity.go).
// An admin can still open an archived table from /admin/games.

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestArchivedTableRefusesEveryWayIn(t *testing.T) {
	l := newTestLobby(t)
	meta, _ := l.Create("Retired")
	_, alice, err := l.Join(meta.ID, meta.InviteToken, "Alice")
	if err != nil {
		t.Fatal(err)
	}
	_, bob, err := l.Join(meta.ID, meta.InviteToken, "Bob")
	if err != nil {
		t.Fatal(err)
	}
	uploadDummyDeck(t, l, meta.ID, alice)
	uploadDummyDeck(t, l, meta.ID, bob)
	if _, err := l.SetArchived(meta.ID, true); err != nil {
		t.Fatal(err)
	}

	if _, _, err := l.Join(meta.ID, meta.InviteToken, "Carol"); !errors.Is(err, ErrGameArchived) {
		t.Errorf("Join: %v, want ErrGameArchived", err)
	}
	if _, _, err := l.JoinAgent(meta.ID, meta.InviteToken, "Agent", AgentDecl{Client: "test"}); !errors.Is(err, ErrGameArchived) {
		t.Errorf("JoinAgent: %v, want ErrGameArchived", err)
	}
	if _, err := l.Spectate(meta.ID, meta.SpectatorInvite); !errors.Is(err, ErrGameArchived) {
		t.Errorf("Spectate: %v, want ErrGameArchived", err)
	}
	cards := []game.Card{game.NewCommander("Other Commander", uuid.Nil), game.NewCard("Other Filler", uuid.Nil)}
	if _, err := l.SetDeck(meta.ID, alice, "other", cards); !errors.Is(err, ErrGameArchived) {
		t.Errorf("SetDeck: %v, want ErrGameArchived", err)
	}
	if _, _, err := l.AddBot(meta.ID, "Bot 1", "random", "", "Mono Red", botDeck(20)); !errors.Is(err, ErrGameArchived) {
		t.Errorf("AddBot: %v, want ErrGameArchived", err)
	}
	if _, err := l.Start(meta.ID); !errors.Is(err, ErrGameArchived) {
		t.Errorf("Start: %v, want ErrGameArchived", err)
	}
	got, _ := l.Get(meta.ID)
	if len(got.Players) != 2 || got.State != string(game.StateLobby) {
		t.Fatalf("archived table changed: %d seats, state %q", len(got.Players), got.State)
	}

	// The refusals were the archive and nothing else: unarchived, the
	// same table takes a seat and starts.
	if _, err := l.SetArchived(meta.ID, false); err != nil {
		t.Fatal(err)
	}
	if _, err := l.Spectate(meta.ID, meta.SpectatorInvite); err != nil {
		t.Errorf("Spectate after unarchive: %v", err)
	}
	if _, err := l.Start(meta.ID); err != nil {
		t.Errorf("Start after unarchive: %v", err)
	}
}

// TestArchivedTableRefusesOverHTTPAndWebSocket is the same refusal as a
// player sees it: the join, spectate and start routes answer 409 with a
// readable message, and a session minted before the archive cannot
// reconnect its socket. An admin still can.
func TestArchivedTableRefusesOverHTTPAndWebSocket(t *testing.T) {
	srv, l, a := newTestHTTPStack(t)
	meta, _ := l.Create("Retired")
	aliceTok, _ := mustJoin(t, srv, meta, "Alice")
	watchTok, err := spectateOverHTTP(srv, meta)
	if err != nil {
		t.Fatal(err)
	}
	// Both sessions connect while the table is live.
	for name, tok := range map[string]string{"player": aliceTok, "spectator": watchTok} {
		if status := dialStatus(t, srv, tok, ""); status != http.StatusSwitchingProtocols {
			t.Fatalf("%s socket before archive: %d", name, status)
		}
	}

	adminTok := adminSession(t, a)
	resp := postJSON(t, srv, "/games/"+meta.ID.String()+"/archive", adminTok, nil)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("archive: %d", resp.StatusCode)
	}

	refused := func(what string, resp *http.Response) {
		t.Helper()
		defer resp.Body.Close()
		var body struct {
			Error string `json:"error"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&body)
		if resp.StatusCode != http.StatusConflict || !strings.Contains(body.Error, "archived") {
			t.Errorf("%s: %d %q, want 409 saying the table is archived", what, resp.StatusCode, body.Error)
		}
	}
	base := "/games/" + meta.ID.String()
	refused("join", postJSON(t, srv, base+"/join", "", joinRequest{InviteToken: meta.InviteToken, Name: "Carol"}))
	refused("spectate", postJSON(t, srv, base+"/spectate", "", spectateRequest{InviteToken: meta.SpectatorInvite, Name: "w"}))
	refused("start", postJSON(t, srv, base+"/start", aliceTok, nil))

	for name, tok := range map[string]string{"player": aliceTok, "spectator": watchTok} {
		if status := dialStatus(t, srv, tok, ""); status != http.StatusConflict {
			t.Errorf("%s socket after archive: %d, want 409", name, status)
		}
	}
	if status := dialStatus(t, srv, adminTok, "game="+meta.ID.String()); status != http.StatusSwitchingProtocols {
		t.Errorf("admin socket on an archived table: %d, want 101", status)
	}

	// Unarchived, the old session reconnects again.
	if _, err := l.SetArchived(meta.ID, false); err != nil {
		t.Fatal(err)
	}
	if status := dialStatus(t, srv, aliceTok, ""); status != http.StatusSwitchingProtocols {
		t.Errorf("player socket after unarchive: %d, want 101", status)
	}
}
