package lobby

import (
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

func TestStartRequiresMinPlayers(t *testing.T) {
	l := newTestLobby(t)
	meta, _ := l.Create("FNM")
	_, _, _ = l.Join(meta.ID, meta.InviteToken, "Solo")
	if _, err := l.Start(meta.ID); err != game.ErrNotEnoughPlayers {
		t.Errorf("Start with 1 player: got %v, want ErrNotEnoughPlayers", err)
	}
}

func TestStartHappyPath(t *testing.T) {
	l := newTestLobby(t)
	meta, _ := l.Create("FNM")
	_, _, _ = l.Join(meta.ID, meta.InviteToken, "Alice")
	_, _, _ = l.Join(meta.ID, meta.InviteToken, "Bob")
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
	_, _, _ = l.Join(meta.ID, meta.InviteToken, "Alice")
	_, _, _ = l.Join(meta.ID, meta.InviteToken, "Bob")
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

func TestGetUnknownGame(t *testing.T) {
	l := newTestLobby(t)
	if _, err := l.Get(uuid.New()); err != ErrGameNotFound {
		t.Errorf("Get unknown: got %v, want ErrGameNotFound", err)
	}
}
