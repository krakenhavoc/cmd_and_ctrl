package lobby

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

func newTestLobby(t *testing.T) *Lobby {
	t.Helper()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	mgr := ws.NewRoomManager(log, "")
	return NewLobby(mgr)
}

func TestCreateGame(t *testing.T) {
	l := newTestLobby(t)
	meta, err := l.Create("Friday Night")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if meta.ID == uuid.Nil {
		t.Error("Create: game ID is nil")
	}
	if meta.Name != "Friday Night" {
		t.Errorf("name: got %q, want %q", meta.Name, "Friday Night")
	}
	if meta.InviteToken == "" {
		t.Error("Create: invite token is empty")
	}
	if meta.State != string(game.StateLobby) {
		t.Errorf("state: got %q, want %q", meta.State, game.StateLobby)
	}
	if len(meta.Players) != 0 {
		t.Errorf("players: got %d, want 0", len(meta.Players))
	}
}

func TestCreateRejectsEmptyName(t *testing.T) {
	l := newTestLobby(t)
	if _, err := l.Create("   "); err != ErrEmptyName {
		t.Errorf("Create empty name: got %v, want ErrEmptyName", err)
	}
}

func TestJoinGameHappyPath(t *testing.T) {
	l := newTestLobby(t)
	meta, err := l.Create("FNM")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	after, playerID, err := l.Join(meta.ID, meta.InviteToken, "Alice")
	if err != nil {
		t.Fatalf("Join: %v", err)
	}
	if playerID == uuid.Nil {
		t.Error("Join: playerID is nil")
	}
	if len(after.Players) != 1 {
		t.Fatalf("players: got %d, want 1", len(after.Players))
	}
	if after.Players[0].Name != "Alice" || after.Players[0].PlayerID != playerID {
		t.Errorf("seat 0: got %+v, want (Alice, %s)", after.Players[0], playerID)
	}
}

func TestJoinRejectsBadInvite(t *testing.T) {
	l := newTestLobby(t)
	meta, _ := l.Create("FNM")
	if _, _, err := l.Join(meta.ID, "not-the-invite", "Alice"); err != ErrInvalidInvite {
		t.Errorf("Join bad invite: got %v, want ErrInvalidInvite", err)
	}
}

func TestJoinRejectsUnknownGame(t *testing.T) {
	l := newTestLobby(t)
	if _, _, err := l.Join(uuid.New(), "anything", "Alice"); err != ErrGameNotFound {
		t.Errorf("Join unknown game: got %v, want ErrGameNotFound", err)
	}
}

func TestJoinRejectsEmptyName(t *testing.T) {
	l := newTestLobby(t)
	meta, _ := l.Create("FNM")
	if _, _, err := l.Join(meta.ID, meta.InviteToken, "   "); err != ErrEmptyName {
		t.Errorf("Join empty name: got %v, want ErrEmptyName", err)
	}
}

func TestJoinFullGame(t *testing.T) {
	l := newTestLobby(t)
	meta, _ := l.Create("FNM")
	for i := 0; i < game.MaxPlayers; i++ {
		if _, _, err := l.Join(meta.ID, meta.InviteToken, "P"+string(rune('A'+i))); err != nil {
			t.Fatalf("Join %d: %v", i, err)
		}
	}
	if _, _, err := l.Join(meta.ID, meta.InviteToken, "One too many"); err != ErrGameFull {
		t.Errorf("Join over capacity: got %v, want ErrGameFull", err)
	}
}

// uploadDummyDeck gives a seat a minimal valid deck so Start can
// transition the game. Tests that exercise Start lifecycle (not
// deck validation) use this to skip the deck-upload flow. Uses
// game.Card directly rather than going through the deck package so
// the lobby test stays hermetic.
func uploadDummyDeck(t *testing.T, l *Lobby, gameID, playerID uuid.UUID) {
	t.Helper()
	cmd := game.NewCommander("Dummy Commander", uuid.Nil)
	filler := game.NewCard("Dummy Filler", uuid.Nil)
	if _, err := l.SetDeck(gameID, playerID, "dummy", []game.Card{cmd, filler}); err != nil {
		t.Fatalf("SetDeck: %v", err)
	}
}

func TestStartRequiresMinPlayers(t *testing.T) {
	l := newTestLobby(t)
	meta, _ := l.Create("FNM")
	_, pid, _ := l.Join(meta.ID, meta.InviteToken, "Solo")
	uploadDummyDeck(t, l, meta.ID, pid)
	if _, err := l.Start(meta.ID); err != game.ErrNotEnoughPlayers {
		t.Errorf("Start with 1 player: got %v, want ErrNotEnoughPlayers", err)
	}
}

func TestStartRequiresAllDecks(t *testing.T) {
	l := newTestLobby(t)
	meta, _ := l.Create("FNM")
	_, alice, _ := l.Join(meta.ID, meta.InviteToken, "Alice")
	_, _, _ = l.Join(meta.ID, meta.InviteToken, "Bob")
	uploadDummyDeck(t, l, meta.ID, alice)
	// Bob hasn't uploaded — Start should refuse.
	if _, err := l.Start(meta.ID); err != ErrDeckNotUploaded {
		t.Errorf("Start with one missing deck: got %v, want ErrDeckNotUploaded", err)
	}
}

func TestStartHappyPath(t *testing.T) {
	l := newTestLobby(t)
	meta, _ := l.Create("FNM")
	_, alice, _ := l.Join(meta.ID, meta.InviteToken, "Alice")
	_, bob, _ := l.Join(meta.ID, meta.InviteToken, "Bob")
	uploadDummyDeck(t, l, meta.ID, alice)
	uploadDummyDeck(t, l, meta.ID, bob)
	after, err := l.Start(meta.ID)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	if after.State != string(game.StateActive) {
		t.Errorf("state: got %q, want %q", after.State, game.StateActive)
	}
	// Idempotent: second Start is a no-op.
	again, err := l.Start(meta.ID)
	if err != nil {
		t.Errorf("Start idempotent: got %v, want nil", err)
	}
	if again.State != string(game.StateActive) {
		t.Errorf("state after re-Start: got %q, want %q", again.State, game.StateActive)
	}
}

func TestJoinAfterStartRejected(t *testing.T) {
	l := newTestLobby(t)
	meta, _ := l.Create("FNM")
	_, alice, _ := l.Join(meta.ID, meta.InviteToken, "Alice")
	_, bob, _ := l.Join(meta.ID, meta.InviteToken, "Bob")
	uploadDummyDeck(t, l, meta.ID, alice)
	uploadDummyDeck(t, l, meta.ID, bob)
	_, _ = l.Start(meta.ID)
	if _, _, err := l.Join(meta.ID, meta.InviteToken, "Carol"); err != ErrGameStarted {
		t.Errorf("Join after Start: got %v, want ErrGameStarted", err)
	}
}

func TestListStripsInviteToken(t *testing.T) {
	l := newTestLobby(t)
	a, _ := l.Create("A")
	b, _ := l.Create("B")

	list := l.List()
	if len(list) != 2 {
		t.Fatalf("List len: got %d, want 2", len(list))
	}
	for _, g := range list {
		if g.InviteToken != "" {
			t.Errorf("List must strip invite token for game %q, got %q", g.Name, g.InviteToken)
		}
	}
	// Get returns the invite token.
	got, err := l.Get(a.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.InviteToken == "" {
		t.Error("Get must include invite token")
	}
	_ = b
}

// TestListGamesPlayersNonNil is a regression for an on-wire bug where
// empty Players slices marshalled as JSON `null` instead of `[]`.
// The Svelte client iterates `g.players` and accessing `.length` on
// null threw mid-render, which Svelte silently caught — leaving the
// stale "no games yet" fallback visible after every create.
func TestListGamesPlayersNonNil(t *testing.T) {
	l := newTestLobby(t)
	_, _ = l.Create("A")

	for _, g := range l.List() {
		if g.Players == nil {
			t.Fatalf("List: Players slice is nil for %q; must be non-nil []SeatInfo{}", g.Name)
		}
	}

	// And make sure the JSON wire renders as `[]`, not `null`. Tests
	// the full round-trip the HTTP handler actually emits.
	raw, err := json.Marshal(l.List())
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if bytes.Contains(raw, []byte(`"players":null`)) {
		t.Errorf("wire format contains players:null, want players:[]; got %s", raw)
	}
	if !bytes.Contains(raw, []byte(`"players":[]`)) {
		t.Errorf("wire format missing players:[]; got %s", raw)
	}

	// Get() goes through copyMeta too.
	got, _ := l.Get(l.List()[0].ID)
	if got.Players == nil {
		t.Error("Get: Players slice is nil")
	}
}

func TestGetUnknownGame(t *testing.T) {
	l := newTestLobby(t)
	if _, err := l.Get(uuid.New()); err != ErrGameNotFound {
		t.Errorf("Get unknown: got %v, want ErrGameNotFound", err)
	}
}
