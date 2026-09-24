package lobby

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
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
	started, err := l.LookupGame(meta.ID)
	if err != nil {
		t.Fatalf("LookupGame: %v", err)
	}
	view := protocol.ViewOfGame(started)
	if view.Turn.ActiveSeat != view.StartingSeat {
		t.Errorf("active seat %d, rolled starting seat %d", view.Turn.ActiveSeat, view.StartingSeat)
	}
	rolls := 0
	for _, entry := range view.Log {
		if entry.Kind == protocol.LogRoll && entry.Sides == 20 && entry.Turn == 0 {
			rolls++
		}
	}
	if rolls < 2 {
		t.Errorf("pregame d20 log entries = %d, want at least one per seat", rolls)
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

// recordingBroadcaster captures BroadcastState calls so tests can
// assert lobby HTTP mutations reach connected clients (the bug this
// guards: Game.Start used to bypass the room, so players already on
// the game page never saw the game start).
type recordingBroadcaster struct {
	calls []struct {
		gameID uuid.UUID
		seq    uint64
		state  string
	}
}

func (r *recordingBroadcaster) BroadcastState(gameID uuid.UUID, seq uint64, view protocol.GameView) {
	r.calls = append(r.calls, struct {
		gameID uuid.UUID
		seq    uint64
		state  string
	}{gameID, seq, view.State})
}

func TestLobbyMutationsBroadcastToRoom(t *testing.T) {
	l := newTestLobby(t)
	rec := &recordingBroadcaster{}
	l.SetStateBroadcaster(rec)

	meta, err := l.Create("FNM")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	_, p1, err := l.Join(meta.ID, meta.InviteToken, "Alice")
	if err != nil {
		t.Fatalf("Join Alice: %v", err)
	}
	_, p2, err := l.Join(meta.ID, meta.InviteToken, "Bob")
	if err != nil {
		t.Fatalf("Join Bob: %v", err)
	}
	deck := []game.Card{
		game.NewCommander("Cmdr", uuid.Nil),
		game.NewCard("Filler", uuid.Nil),
	}
	if _, err := l.SetDeck(meta.ID, p1, "Deck A", deck); err != nil {
		t.Fatalf("SetDeck p1: %v", err)
	}
	if _, err := l.SetDeck(meta.ID, p2, "Deck B", deck); err != nil {
		t.Fatalf("SetDeck p2: %v", err)
	}
	if _, err := l.Start(meta.ID); err != nil {
		t.Fatalf("Start: %v", err)
	}

	// 2 joins + 2 deck uploads + 1 start, every one broadcast with a
	// strictly increasing seq, all for this game, ending active.
	if len(rec.calls) != 5 {
		t.Fatalf("broadcasts: got %d, want 5 (%+v)", len(rec.calls), rec.calls)
	}
	for i, c := range rec.calls {
		if c.gameID != meta.ID {
			t.Errorf("call %d: gameID %s, want %s", i, c.gameID, meta.ID)
		}
		if i > 0 && c.seq <= rec.calls[i-1].seq {
			t.Errorf("call %d: seq %d not above previous %d", i, c.seq, rec.calls[i-1].seq)
		}
	}
	if got := rec.calls[4].state; got != string(game.StateActive) {
		t.Errorf("final broadcast state: got %q, want %q", got, game.StateActive)
	}

	// A failed mutation must not broadcast.
	if _, _, err := l.Join(meta.ID, meta.InviteToken, "Carol"); err == nil {
		t.Fatal("Join after start: expected error")
	}
	if len(rec.calls) != 5 {
		t.Fatalf("failed join broadcast: got %d calls, want still 5", len(rec.calls))
	}
}
