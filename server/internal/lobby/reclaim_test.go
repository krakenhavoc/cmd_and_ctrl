package lobby

// reclaim_test.go is an auth-surface test file. Most of it is about
// what a reclaim link must REFUSE to do; the happy path is one test
// and the rest are the reasons this is safe to hand to a friend over
// Discord.

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// The headline: a player who lost their session gets their OWN seat
// back in a game that has already started, and the session works on
// the WebSocket with that player's hidden information.
func TestReclaimReturnsAPlayerToTheirSeatInAStartedGame(t *testing.T) {
	srv, l, a := newTestHTTPStack(t)
	meta, alice, _ := startTwoSeatGame(t, l, "FNM")

	// The state the feature exists for: the invite link is useless
	// once the table is underway.
	if _, _, err := l.Join(meta.ID, meta.InviteToken, "Alice"); err != ErrGameStarted {
		t.Fatalf("Join on a started game: %v, want ErrGameStarted", err)
	}

	ticket := mintTicket(t, srv, adminSession(t, a), meta.ID, alice)
	if ticket.Ticket == "" {
		t.Fatal("empty ticket")
	}
	if !ticket.SingleUse {
		t.Error("response does not advertise single use")
	}
	if ticket.TTLSeconds != int(ReclaimTTL/time.Second) {
		t.Errorf("ttl_seconds = %d, want %d", ticket.TTLSeconds, int(ReclaimTTL/time.Second))
	}
	if ticket.ExpiresAt.After(time.Now().Add(ReclaimTTL + time.Minute)) {
		t.Errorf("expires_at is not short-lived: %v", ticket.ExpiresAt)
	}
	if ticket.PlayerID != alice || ticket.PlayerName != "Alice" {
		t.Errorf("ticket describes the wrong seat: %+v", ticket)
	}

	// Redeem it — no session, exactly like a player who cleared
	// localStorage.
	sess := redeem(t, srv, meta.ID, ticket.Ticket, http.StatusOK)
	if sess.Principal.Role != auth.RolePlayer {
		t.Errorf("role = %q, want player", sess.Principal.Role)
	}
	if sess.PlayerID != alice || sess.Principal.PlayerID != alice {
		t.Errorf("reclaimed the wrong seat: %v, want %v", sess.PlayerID, alice)
	}
	if sess.Principal.GameID != meta.ID {
		t.Errorf("session bound to %v, want %v", sess.Principal.GameID, meta.ID)
	}
	if sess.Game == nil || sess.Game.InviteToken != "" || sess.Game.SpectatorInvite != "" {
		t.Error("reclaim response leaked a table invite")
	}

	// And it actually plays: the WS upgrade binds to Alice's seat and
	// the action gate lets her act — a spectator session would be
	// refused here, which is exactly the outcome this feature exists
	// to avoid.
	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws?token=" + sess.Token
	conn, upgrade, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		body := ""
		if upgrade != nil {
			b, _ := io.ReadAll(upgrade.Body)
			body = string(b)
		}
		t.Fatalf("ws dial with a reclaimed session: %v (%s)", err, body)
	}
	defer conn.Close()

	_, raw, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read snapshot: %v", err)
	}
	var first struct {
		Kind string `json:"kind"`
	}
	if err := json.Unmarshal(raw, &first); err != nil {
		t.Fatalf("unmarshal snapshot: %v", err)
	}
	if first.Kind != "snapshot" {
		t.Fatalf("first frame kind = %q, want snapshot", first.Kind)
	}

	// The payload names Alice explicitly, so the server's
	// caller-vs-player check (ErrPlayerCallerMismatch) is what proves
	// the socket is bound to HER seat and not merely to the table.
	action := []byte(`{"v":0,"kind":"action","id":"00000000-0000-4000-8000-000000000001",` +
		`"payload":{"type":"draw_card","player":"` + alice.String() + `"}}`)
	if err := conn.WriteMessage(websocket.TextMessage, action); err != nil {
		t.Fatalf("write action: %v", err)
	}
	_, raw, err = conn.ReadMessage()
	if err != nil {
		t.Fatalf("read action response: %v", err)
	}
	var reply struct {
		Kind    string `json:"kind"`
		Payload struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"payload"`
	}
	_ = json.Unmarshal(raw, &reply)
	if reply.Kind == "error" {
		t.Fatalf("reclaimed seat could not act: %s / %s", reply.Payload.Code, reply.Payload.Message)
	}
}

// Single use is the property that makes a link forwarded in a chat
// safe-ish to forget about.
func TestReclaimTicketIsSingleUse(t *testing.T) {
	srv, l, a := newTestHTTPStack(t)
	meta, alice, _ := startTwoSeatGame(t, l, "FNM")
	ticket := mintTicket(t, srv, adminSession(t, a), meta.ID, alice)

	redeem(t, srv, meta.ID, ticket.Ticket, http.StatusOK)
	redeem(t, srv, meta.ID, ticket.Ticket, http.StatusUnauthorized)
}

func TestReclaimTicketExpires(t *testing.T) {
	l := newTestLobby(t)
	meta, alice, _ := startTwoSeatGame(t, l, "FNM")

	ticket, err := l.MintReclaim(meta.ID, alice)
	if err != nil {
		t.Fatalf("MintReclaim: %v", err)
	}
	// Reach into the store rather than sleeping 15 minutes.
	l.mu.Lock()
	key := reclaimKey(ticket.Token)
	rec := l.reclaims[key]
	rec.expiresAt = time.Now().UTC().Add(-time.Second)
	l.reclaims[key] = rec
	l.mu.Unlock()

	if _, _, err := l.RedeemReclaim(meta.ID, ticket.Token); err != ErrInvalidReclaim {
		t.Errorf("expired ticket: %v, want ErrInvalidReclaim", err)
	}
}

// Forged, wrong-table and empty tickets all get the same answer, so
// the response cannot be used to tell them apart.
func TestReclaimRefusesForgedAndCrossTableTickets(t *testing.T) {
	l := newTestLobby(t)
	metaA, alice, _ := startTwoSeatGame(t, l, "Table A")
	metaB, _, _ := startTwoSeatGame(t, l, "Table B")

	ticket, err := l.MintReclaim(metaA.ID, alice)
	if err != nil {
		t.Fatal(err)
	}
	for name, tc := range map[string]struct {
		game uuid.UUID
		tok  string
	}{
		"empty":              {metaA.ID, ""},
		"garbage":            {metaA.ID, "not-a-ticket"},
		"invite token":       {metaA.ID, metaA.InviteToken},
		"spectator invite":   {metaA.ID, metaA.SpectatorInvite},
		"game id":            {metaA.ID, metaA.ID.String()},
		"player id":          {metaA.ID, alice.String()},
		"right tok, table B": {metaB.ID, ticket.Token},
	} {
		if _, _, err := l.RedeemReclaim(tc.game, tc.tok); err != ErrInvalidReclaim {
			t.Errorf("%s: %v, want ErrInvalidReclaim", name, err)
		}
	}
	// The genuine one still works afterwards: the failed attempts
	// did not consume it.
	if _, _, err := l.RedeemReclaim(metaA.ID, ticket.Token); err != nil {
		t.Errorf("genuine ticket after failed attempts: %v", err)
	}
}

// A reclaim link returns an EXISTING seat. It is never a way to add
// one, and never a way to take a bot's chair.
func TestMintRefusesSeatsThatAreNotADisconnectedPlayer(t *testing.T) {
	l := newTestLobby(t)
	host := newFakeBotHost()
	l.SetBotHost(host)

	meta, _ := l.Create("FNM")
	_, human, err := l.Join(meta.ID, meta.InviteToken, "Alice")
	if err != nil {
		t.Fatal(err)
	}
	_, botID, err := l.AddBot(meta.ID, "Bot 1", "random", "", "Mono Red", botDeck(20))
	if err != nil {
		t.Fatal(err)
	}

	if _, err := l.MintReclaim(meta.ID, uuid.New()); err != ErrPlayerNotInGame {
		t.Errorf("mint for a stranger: %v, want ErrPlayerNotInGame", err)
	}
	if _, err := l.MintReclaim(uuid.New(), human); err != ErrGameNotFound {
		t.Errorf("mint for an unknown table: %v, want ErrGameNotFound", err)
	}
	if _, err := l.MintReclaim(meta.ID, botID); err != ErrSeatIsBot {
		t.Errorf("mint for a bot seat: %v, want ErrSeatIsBot", err)
	}
}

// The seat is re-checked at redemption, not trusted from mint time.
func TestReclaimRefusesASeatThatLeftTheTable(t *testing.T) {
	l := newTestLobby(t)
	host := newFakeBotHost()
	l.SetBotHost(host)

	meta, _ := l.Create("FNM")
	if _, _, err := l.Join(meta.ID, meta.InviteToken, "Alice"); err != nil {
		t.Fatal(err)
	}
	_, botID, err := l.AddBot(meta.ID, "Bot 1", "random", "", "Mono Red", botDeck(20))
	if err != nil {
		t.Fatal(err)
	}
	// Mint against a human seat, then make that seat vanish. RemoveBot
	// is the only unseat path today, so drive it through the bot seat
	// with the bot flag temporarily cleared — the point under test is
	// RedeemReclaim's re-check, not how the seat left.
	ticket, err := l.MintReclaim(meta.ID, botIDAsHuman(t, l, meta.ID, botID))
	if err != nil {
		t.Fatalf("MintReclaim: %v", err)
	}
	restoreBotFlag(t, l, meta.ID, botID)
	if _, err := l.RemoveBot(meta.ID, botID); err != nil {
		t.Fatalf("RemoveBot: %v", err)
	}
	if _, _, err := l.RedeemReclaim(meta.ID, ticket.Token); err != ErrPlayerNotInGame {
		t.Errorf("redeem for a seat that left: %v, want ErrPlayerNotInGame", err)
	}
}

func TestReclaimRefusesArchivedTables(t *testing.T) {
	l := newTestLobby(t)
	meta, alice, _ := startTwoSeatGame(t, l, "FNM")

	ticket, err := l.MintReclaim(meta.ID, alice)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := l.SetArchived(meta.ID, true); err != nil {
		t.Fatal(err)
	}
	// Archiving drops outstanding tickets outright, so the one minted
	// a moment ago is simply not there any more.
	if _, _, err := l.RedeemReclaim(meta.ID, ticket.Token); err != ErrInvalidReclaim {
		t.Errorf("redeem against an archived table: %v, want ErrInvalidReclaim", err)
	}
	if _, err := l.MintReclaim(meta.ID, alice); err != ErrGameArchived {
		t.Errorf("mint for an archived table: %v, want ErrGameArchived", err)
	}
}

func TestDeletingATableDropsItsReclaimTickets(t *testing.T) {
	l := newTestLobby(t)
	meta, alice, _ := startTwoSeatGame(t, l, "FNM")
	ticket, err := l.MintReclaim(meta.ID, alice)
	if err != nil {
		t.Fatal(err)
	}
	if err := l.Delete(meta.ID); err != nil {
		t.Fatal(err)
	}
	if _, _, err := l.RedeemReclaim(meta.ID, ticket.Token); err != ErrInvalidReclaim {
		t.Errorf("redeem after delete: %v, want ErrInvalidReclaim", err)
	}
	l.mu.Lock()
	outstanding := len(l.reclaims)
	l.mu.Unlock()
	if outstanding != 0 {
		t.Errorf("%d tickets still held for a deleted table", outstanding)
	}
}

// Minting is the admin's job and nobody else's — a seated player
// must not be able to mint a credential for the seat next to them.
func TestMintReclaimIsAdminOnly(t *testing.T) {
	srv, l, a := newTestHTTPStack(t)
	meta, alice, bob := startTwoSeatGame(t, l, "FNM")

	path := "/games/" + meta.ID.String() + "/seats/" + alice.String() + "/reclaim"

	// Bob is seated at this very table and still cannot mint.
	bobTok, _, err := a.Issue(context.Background(), auth.Principal{
		Role: auth.RolePlayer, GameID: meta.ID, PlayerID: bob, Name: "Bob",
	}, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	resp := postJSON(t, srv, path, bobTok, nil)
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("seated player mint: got %d, want %d", resp.StatusCode, http.StatusForbidden)
	}
	resp.Body.Close()

	// Nor can an anonymous caller.
	resp = postJSON(t, srv, path, "", nil)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("anonymous mint: got %d, want %d", resp.StatusCode, http.StatusUnauthorized)
	}
	resp.Body.Close()

	// The admin can.
	resp = postJSON(t, srv, path, adminSession(t, a), nil)
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("admin mint: got %d, want %d", resp.StatusCode, http.StatusCreated)
	}
	resp.Body.Close()
}

// The ticket is a secret and lives in exactly one place: the mint
// response. Not the metadata, not the disk.
func TestReclaimTicketIsNeverPersistedOrEchoedInMetadata(t *testing.T) {
	dir := t.TempDir()
	log := quietLogger()
	mgr := ws.NewRoomManager(log, dir)
	l := NewLobby(mgr)

	meta, alice, _ := startTwoSeatGame(t, l, "FNM")
	ticket, err := l.MintReclaim(meta.ID, alice)
	if err != nil {
		t.Fatal(err)
	}

	got, err := l.Get(meta.ID)
	if err != nil {
		t.Fatal(err)
	}
	blob, _ := json.Marshal(got)
	if strings.Contains(string(blob), ticket.Token) {
		t.Error("GameMeta carries the reclaim ticket")
	}
	raw, err := os.ReadFile(metaPath(dir, meta.ID))
	if err != nil {
		t.Fatalf("read persisted meta: %v", err)
	}
	if strings.Contains(string(raw), ticket.Token) {
		t.Error("reclaim ticket was written to disk")
	}

	// It also dies with the process: the store is in memory on
	// purpose (see reclaim.go — deploy survival is #517's job).
	mgr2 := ws.NewRoomManager(log, dir)
	l2 := NewLobby(mgr2)
	l2.RestoreFromDisk(log)
	if _, _, err := l2.RedeemReclaim(meta.ID, ticket.Token); err != ErrInvalidReclaim {
		t.Errorf("ticket survived a restart: %v", err)
	}
}

// Every ticket is minted, never derived: two mints for the same seat
// produce different strings and neither contains anything the caller
// already knew.
func TestReclaimTicketsAreMintedNotDerived(t *testing.T) {
	l := newTestLobby(t)
	meta, alice, _ := startTwoSeatGame(t, l, "FNM")

	seen := map[string]bool{}
	for i := 0; i < 8; i++ {
		ticket, err := l.MintReclaim(meta.ID, alice)
		if err != nil {
			t.Fatal(err)
		}
		if seen[ticket.Token] {
			t.Fatal("two mints produced the same ticket")
		}
		seen[ticket.Token] = true
		for _, known := range []string{meta.ID.String(), alice.String(), meta.InviteToken, meta.SpectatorInvite} {
			if strings.Contains(ticket.Token, known) {
				t.Errorf("ticket embeds %q", known)
			}
		}
	}
}

// Reclaim must not have quietly widened the ordinary invite flow:
// a stranger holding the table's invite token still cannot take a
// seat in a started game, and can still spectate.
func TestReclaimDoesNotWidenTheInviteToken(t *testing.T) {
	srv, l, _ := newTestHTTPStack(t)
	meta, _, _ := startTwoSeatGame(t, l, "FNM")

	resp := postJSON(t, srv, "/games/"+meta.ID.String()+"/join", "",
		joinRequest{InviteToken: meta.InviteToken, Name: "Stranger"})
	if resp.StatusCode != http.StatusConflict {
		t.Errorf("join with the invite on a started game: got %d, want %d", resp.StatusCode, http.StatusConflict)
	}
	resp.Body.Close()

	// The invite token is not a reclaim ticket either.
	redeem(t, srv, meta.ID, meta.InviteToken, http.StatusUnauthorized)

	// Spectating still works exactly as before.
	resp = postJSON(t, srv, "/games/"+meta.ID.String()+"/spectate", "",
		spectateRequest{InviteToken: meta.SpectatorInvite, Name: "watcher"})
	if resp.StatusCode != http.StatusOK {
		t.Errorf("spectate: got %d, want %d", resp.StatusCode, http.StatusOK)
	}
	resp.Body.Close()
}

// --- helpers -------------------------------------------------------

func mintTicket(t *testing.T, srv *httptest.Server, adminTok string, gameID, playerID uuid.UUID) reclaimTicketResponse {
	t.Helper()
	resp := postJSON(t, srv, "/games/"+gameID.String()+"/seats/"+playerID.String()+"/reclaim", adminTok, nil)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("mint: got %d (%s)", resp.StatusCode, b)
	}
	var out reclaimTicketResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode ticket: %v", err)
	}
	return out
}

func redeem(t *testing.T, srv *httptest.Server, gameID uuid.UUID, ticket string, want int) sessionResponse {
	t.Helper()
	resp := postJSON(t, srv, "/games/"+gameID.String()+"/reclaim", "", reclaimRequest{Ticket: ticket})
	defer resp.Body.Close()
	if resp.StatusCode != want {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("redeem: got %d, want %d (%s)", resp.StatusCode, want, b)
	}
	var out sessionResponse
	if want == http.StatusOK {
		if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
			t.Fatalf("decode session: %v", err)
		}
	}
	return out
}

// botIDAsHuman temporarily clears the seat's bot flag so MintReclaim
// will issue for it, and returns the player ID. Paired with
// restoreBotFlag.
func botIDAsHuman(t *testing.T, l *Lobby, gameID, playerID uuid.UUID) uuid.UUID {
	t.Helper()
	l.mu.Lock()
	defer l.mu.Unlock()
	e := l.games[gameID]
	for i := range e.meta.Players {
		if e.meta.Players[i].PlayerID == playerID {
			e.meta.Players[i].IsBot = false
			return playerID
		}
	}
	t.Fatalf("seat %v not found", playerID)
	return uuid.Nil
}

func restoreBotFlag(t *testing.T, l *Lobby, gameID, playerID uuid.UUID) {
	t.Helper()
	l.mu.Lock()
	defer l.mu.Unlock()
	e := l.games[gameID]
	for i := range e.meta.Players {
		if e.meta.Players[i].PlayerID == playerID {
			e.meta.Players[i].IsBot = true
			return
		}
	}
}
