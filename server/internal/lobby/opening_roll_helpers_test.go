package lobby

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// finishOpeningRoll plays a started table's opening roll out the way
// the admin would (ADR 0121 §1): "Roll for everyone left" until one
// leader remains, then that leader takes the first turn. Since the
// lobby's Start opens the roll rather than playing it (ADR 0121 PR 5),
// a test that needs dealt hands or a turn to act in calls this after
// Start. It goes through the room, so each step is applied the way the
// hub applies it: without an undo entry. It returns how many steps it
// applied (each is one replay line).
func finishOpeningRoll(t *testing.T, l *Lobby, id uuid.UUID) int {
	t.Helper()
	room := l.RoomOf(id)
	if room == nil {
		t.Fatalf("finishOpeningRoll: no room for %s", id)
	}
	// Every round shrinks to the tied leaders of the one before, and a
	// d20 tie is rare, so this many steps is generous.
	for i := 0; i < 64; i++ {
		v := protocol.ViewOfGame(room.Game)
		if v.OpeningRoll == nil {
			return i
		}
		var step func() error
		if c := v.OpeningRoll.Chooser; c != nil {
			seat := *c
			chooser, err := uuid.Parse(v.Seats[seat].ID)
			if err != nil {
				t.Fatalf("finishOpeningRoll: chooser id: %v", err)
			}
			step = func() error { return room.Game.ChooseStartingPlayer(chooser, seat) }
		} else {
			step = func() error { return room.Game.HostRollRemaining(uuid.Nil) }
		}
		if _, _, err := room.ApplyExternal(step); err != nil {
			t.Fatalf("finishOpeningRoll: %v", err)
		}
	}
	t.Fatal("finishOpeningRoll: the opening roll did not finish")
	return 0
}
