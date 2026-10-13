package lobby

// blocks_test.go: remove-and-block and the name filter at join (#2984).

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

func newTestRoomManager(t *testing.T) *ws.RoomManager {
	t.Helper()
	return ws.NewRoomManager(quietLogger(), "")
}

// blockedTable seats Alice (host), Bob (a guest) and returns the lobby.
func blockedTable(t *testing.T) (*Lobby, GameMeta, uuid.UUID) {
	t.Helper()
	l := NewLobby(newTestRoomManager(t))
	meta, err := l.Create("Block table")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := l.Join(meta.ID, meta.InviteToken, "Alice"); err != nil {
		t.Fatal(err)
	}
	_, bob, err := l.Join(meta.ID, meta.InviteToken, "Bob")
	if err != nil {
		t.Fatal(err)
	}
	return l, meta, bob
}

func TestRemoveWithoutBlockLetsThemSitBackDown(t *testing.T) {
	l, meta, bob := blockedTable(t)
	if _, err := l.RemoveSeatBlocking(meta.ID, bob, false); err != nil {
		t.Fatal(err)
	}
	if n := l.BlockCount(meta.ID); n != 0 {
		t.Fatalf("a plain removal recorded %d blocks", n)
	}
	if _, _, err := l.Join(meta.ID, meta.InviteToken, "Bob"); err != nil {
		t.Errorf("rejoin after a plain removal: %v", err)
	}
}

func TestBlockedGuestIsRefusedUnderTheSameNameInAnySpelling(t *testing.T) {
	l, meta, bob := blockedTable(t)
	if _, err := l.RemoveSeatBlocking(meta.ID, bob, true); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"Bob", "bob", "BOB", "b0b", "B O B", "  Bob  ", "Bööb", "bobbb"} {
		_, _, err := l.Join(meta.ID, meta.InviteToken, name)
		if !errors.Is(err, ErrBlockedFromTable) {
			t.Errorf("Join(%q) = %v, want ErrBlockedFromTable", name, err)
		}
	}
	// Everyone else is unaffected, including a guest with another name.
	for _, name := range []string{"Carol", "Rob", "Bobby Tables"} {
		if _, _, err := l.Join(meta.ID, meta.InviteToken, name); err != nil {
			t.Errorf("Join(%q) = %v, want allowed", name, err)
		}
		if _, err := l.Get(meta.ID); err != nil {
			t.Fatal(err)
		}
		// make room for the next one
		m, _ := l.Get(meta.ID)
		if len(m.Players) >= 3 {
			last := m.Players[len(m.Players)-1]
			if _, err := l.RemoveSeat(meta.ID, last.PlayerID); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestBlockIsPerTable(t *testing.T) {
	l, meta, bob := blockedTable(t)
	other, err := l.Create("Another table")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := l.RemoveSeatBlocking(meta.ID, bob, true); err != nil {
		t.Fatal(err)
	}
	if _, _, err := l.Join(other.ID, other.InviteToken, "Bob"); err != nil {
		t.Errorf("a block at one table kept Bob out of another: %v", err)
	}
}

func TestBlockedAccountIsRefusedAndAGuestOfTheSameNameIsNot(t *testing.T) {
	l := NewLobby(newTestRoomManager(t))
	meta, err := l.Create("Account table")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := l.Join(meta.ID, meta.InviteToken, "Alice"); err != nil {
		t.Fatal(err)
	}
	user, other := uuid.New(), uuid.New()
	disc := DiscordIdentity{ID: "9001", Username: "dave_d", GlobalName: "Dave"}
	_, dave, err := l.JoinAs(meta.ID, meta.InviteToken, "", disc, user)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := l.RemoveSeatBlocking(meta.ID, dave, true); err != nil {
		t.Fatal(err)
	}

	// The same account: refused, whatever the display name says.
	if _, _, err := l.JoinAs(meta.ID, meta.InviteToken, "", DiscordIdentity{ID: "9001", Username: "renamed"}, user); !errors.Is(err, ErrBlockedFromTable) {
		t.Errorf("same account: %v, want ErrBlockedFromTable", err)
	}
	// The same Discord id with no users row (a seat claimed before one
	// exists): refused too.
	if _, _, err := l.JoinAs(meta.ID, meta.InviteToken, "", disc, uuid.Nil); !errors.Is(err, ErrBlockedFromTable) {
		t.Errorf("same Discord id: %v, want ErrBlockedFromTable", err)
	}
	// A different person called Dave, signed in or not: allowed. A name
	// block is only ever made against a guest, and an account block holds
	// against the account.
	if _, _, err := l.JoinAs(meta.ID, meta.InviteToken, "", DiscordIdentity{ID: "9002", Username: "other_dave", GlobalName: "Dave"}, other); err != nil {
		t.Errorf("a different account called Dave: %v", err)
	}
	m, _ := l.Get(meta.ID)
	if _, err := l.RemoveSeat(meta.ID, m.Players[len(m.Players)-1].PlayerID); err != nil {
		t.Fatal(err)
	}
	if _, _, err := l.Join(meta.ID, meta.InviteToken, "Dave"); err != nil {
		t.Errorf("a guest called Dave after an ACCOUNT was blocked: %v", err)
	}
}

func TestBlockedGuestCanEvadeWithAnotherName(t *testing.T) {
	// The documented limit (blocks.go): a guest has nothing but a name.
	// Pinned so nobody reads the block as a wall.
	l, meta, bob := blockedTable(t)
	if _, err := l.RemoveSeatBlocking(meta.ID, bob, true); err != nil {
		t.Fatal(err)
	}
	if _, _, err := l.Join(meta.ID, meta.InviteToken, "Robert"); err != nil {
		t.Errorf("a new name was refused: %v", err)
	}
}

func TestBlockSurvivesARestart(t *testing.T) {
	dir := t.TempDir()
	l, _ := newDurableLobby(t, dir)
	meta, err := l.Create("Durable block")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := l.Join(meta.ID, meta.InviteToken, "Alice"); err != nil {
		t.Fatal(err)
	}
	_, bob, err := l.Join(meta.ID, meta.InviteToken, "Bob")
	if err != nil {
		t.Fatal(err)
	}
	_, dave, err := l.JoinWithIdentity(meta.ID, meta.InviteToken, "", DiscordIdentity{ID: "777", Username: "dave"})
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []uuid.UUID{bob, dave} {
		if _, err := l.RemoveSeatBlocking(meta.ID, id, true); err != nil {
			t.Fatal(err)
		}
	}

	// --- the deploy ---
	l2, _ := newDurableLobby(t, dir)
	if n := l2.RestoreFromDisk(quietLogger()); n != 1 {
		t.Fatalf("restored %d games", n)
	}
	if n := l2.BlockCount(meta.ID); n != 2 {
		t.Fatalf("restored table has %d blocks, want 2 (a name and a Discord id)", n)
	}
	if _, _, err := l2.Join(meta.ID, meta.InviteToken, "bob"); !errors.Is(err, ErrBlockedFromTable) {
		t.Errorf("blocked guest after a restart: %v", err)
	}
	if _, _, err := l2.JoinWithIdentity(meta.ID, meta.InviteToken, "", DiscordIdentity{ID: "777", Username: "dave"}); !errors.Is(err, ErrBlockedFromTable) {
		t.Errorf("blocked Discord id after a restart: %v", err)
	}
	if _, _, err := l2.Join(meta.ID, meta.InviteToken, "Carol"); err != nil {
		t.Errorf("an unrelated joiner after a restart: %v", err)
	}
}

func TestDeletingTheTableDeletesItsBlocks(t *testing.T) {
	dir := t.TempDir()
	l, _ := newDurableLobby(t, dir)
	meta, _ := l.Create("Short lived")
	_, _, _ = l.Join(meta.ID, meta.InviteToken, "Alice")
	_, bob, _ := l.Join(meta.ID, meta.InviteToken, "Bob")
	if _, err := l.RemoveSeatBlocking(meta.ID, bob, true); err != nil {
		t.Fatal(err)
	}
	if err := l.Delete(meta.ID); err != nil {
		t.Fatal(err)
	}
	recs, err := l.store.Blocks(t.Context(), meta.ID)
	if err != nil || len(recs) != 0 {
		t.Errorf("blocks of a deleted table: %v, %v", recs, err)
	}
}

// --- the HTTP route -------------------------------------------------------

func TestRemoveRouteBlocksOnRequestAndTheRefusalIsClear(t *testing.T) {
	s := newSeatRemoveStack(t)
	meta, alice, bob := s.table(t)

	resp := doJSON(t, s.srv, http.MethodDelete, "/games/"+meta.ID.String()+"/seats/"+bob.PlayerID.String()+"?block=true", alice.Token, nil)
	s.wantStatus(t, resp, http.StatusOK)

	resp = postJSON(t, s.srv, "/games/"+meta.ID.String()+"/join", "",
		joinRequest{InviteToken: meta.InviteToken, Name: "Bob"})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("blocked rejoin = %d %s, want 403", resp.StatusCode, body)
	}
	var body struct{ Error, Code string }
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.Code != CodeBlockedFromTable || body.Error == "" {
		t.Errorf("body = %+v, want the %q code and a message", body, CodeBlockedFromTable)
	}

	// Another joiner is unaffected.
	joinAs(t, s.srv, meta, "Carol")
}

func TestRemoveRouteWithoutBlockLeavesTheDoorOpen(t *testing.T) {
	s := newSeatRemoveStack(t)
	meta, alice, bob := s.table(t)
	s.wantStatus(t, s.del(t, meta.ID, bob.PlayerID, alice.Token), http.StatusOK)
	joinAs(t, s.srv, meta, "Bob")
}

func TestOnlyWhoMayRemoveMayBlock(t *testing.T) {
	// ?block=true adds nothing to the authorisation: a seat that may not
	// remove may not block.
	s := newSeatRemoveStack(t)
	meta, _, bob := s.table(t)
	carol := joinAs(t, s.srv, meta, "Carol")
	resp := doJSON(t, s.srv, http.MethodDelete, "/games/"+meta.ID.String()+"/seats/"+bob.PlayerID.String()+"?block=true", carol.Token, nil)
	s.wantStatus(t, resp, http.StatusForbidden)
	if n := s.l.BlockCount(meta.ID); n != 0 {
		t.Errorf("a refused request recorded %d blocks", n)
	}
}

// --- the name filter at join and create ------------------------------------

func TestJoinRefusesAListedGuestName(t *testing.T) {
	s := newSeatRemoveStack(t)
	meta, _, _ := s.table(t)
	resp := postJSON(t, s.srv, "/games/"+meta.ID.String()+"/join", "",
		joinRequest{InviteToken: meta.InviteToken, Name: "xX_slût_Xx"})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("status %d %s, want 400", resp.StatusCode, body)
	}
	var body struct{ Error, Code string }
	_ = json.NewDecoder(resp.Body).Decode(&body)
	if body.Code != CodeNameNotAllowed || body.Error == "" {
		t.Errorf("body = %+v", body)
	}
	if n := len(mustGet(t, s.l, meta.ID).Players); n != 2 {
		t.Errorf("a refused name seated someone: %d seats", n)
	}
}

func TestJoinFallsBackToTheDiscordUsernameOrRefuses(t *testing.T) {
	l := NewLobby(newTestRoomManager(t))
	meta, _ := l.Create("Discord names")
	_, p, err := l.JoinWithIdentity(meta.ID, meta.InviteToken, "", DiscordIdentity{ID: "1", Username: "plain_name", GlobalName: "retarded"})
	if err != nil {
		t.Fatalf("a refused global name must fall back to the username: %v", err)
	}
	m, _ := l.Get(meta.ID)
	if got := m.Players[0].Name; got != "plain_name" || m.Players[0].PlayerID != p {
		t.Errorf("seat name = %q, want the username", got)
	}
	if _, _, err := l.JoinWithIdentity(meta.ID, meta.InviteToken, "", DiscordIdentity{ID: "2", Username: "retarded", GlobalName: "kys"}); !errors.Is(err, ErrDiscordNameNotAllowed) {
		t.Errorf("both names listed: %v, want ErrDiscordNameNotAllowed", err)
	}
}

func TestCreateRefusesAListedTableName(t *testing.T) {
	l := NewLobby(newTestRoomManager(t))
	if _, err := l.Create("the sluts of commander"); err == nil {
		t.Fatal("a listed table name was accepted")
	}
	if _, err := l.Create("Friday Night Commander"); err != nil {
		t.Fatal(err)
	}
}

func TestGeneratedTableNameLeavesOutAListedCreatorName(t *testing.T) {
	for i := 0; i < 300; i++ {
		got := SuggestTableName("kill yourself")
		if got == "" || strings.Contains(strings.ToLower(got), "kill") {
			t.Fatalf("generated %q with a refused creator name", got)
		}
	}
}

func TestStartedTableBlockIsAdminOnlyAndRecordedWithTheConcession(t *testing.T) {
	s := newSeatRemoveStack(t)
	meta, seats := s.startedTable(t, uuid.New(), "Alice", "Bob", "Carol")
	alice, bob := seats[0], seats[1]
	path := "/games/" + meta.ID.String() + "/seats/" + bob.PlayerID.String() + "?block=true"

	// The host does not gain the power to kick from a started game by
	// adding a flag: still 409, nothing recorded.
	s.wantStatus(t, doJSON(t, s.srv, http.MethodDelete, path, alice.Token, nil), http.StatusConflict)
	if n := s.l.BlockCount(meta.ID); n != 0 {
		t.Fatalf("a refused request recorded %d blocks", n)
	}

	s.wantStatus(t, doJSON(t, s.srv, http.MethodDelete, path, adminToken(t, s.srv), nil), http.StatusOK)
	if n := s.l.BlockCount(meta.ID); n != 1 {
		t.Errorf("blocks = %d, want Bob's name", n)
	}
	if p := s.enginePlayer(t, meta.ID, bob.PlayerID); !p.Removed {
		t.Error("the seat was not removed")
	}
}
