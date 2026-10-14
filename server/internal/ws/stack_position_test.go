package ws

import (
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// Tests for stack_position.go (#3429): the room keeps each seat's spot,
// stamps every capture with all of them, and the per-seat filter reduces
// the view to the viewer's own.

func TestRoomStampsStackPositionsAndFiltersThemPerSeat(t *testing.T) {
	room := NewRoom(seedTestGame(t), slog.New(slog.NewTextHandler(io.Discard, nil)), "")
	alice, bob := uuid.New(), uuid.New()
	if room.StackPosition(alice) != nil {
		t.Fatal("a new room has a stack position")
	}
	room.SetStackPosition(alice, &protocol.StackPositionView{X: 0.7, Y: 0.2})
	room.SetStackPosition(bob, &protocol.StackPositionView{X: 0.1, Y: 0.9})

	view, _, err := room.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if len(view.StackPositions) != 2 {
		t.Errorf("the full view holds %+v, want both seats", view.StackPositions)
	}
	got := room.ViewFor(alice.String()).StackPositions
	if len(got) != 1 || got[alice.String()] != (protocol.StackPositionView{X: 0.7, Y: 0.2}) {
		t.Errorf("alice's view = %+v, want only her own", got)
	}
	if got := room.ViewFor(protocol.SpectatorViewerID).StackPositions; got != nil {
		t.Errorf("a spectator sees %+v", got)
	}

	room.SetStackPosition(alice, nil)
	if room.StackPosition(alice) != nil || room.ViewFor(alice.String()).StackPositions != nil {
		t.Error("cleared, but alice still has a spot")
	}
	if room.ViewFor(bob.String()).StackPositions == nil {
		t.Error("clearing alice's spot cleared bob's")
	}
}
