package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// initiative_engine_test.go pins the engine half of ADR 0203: the three
// inherent triggers of CR 726.2, CR 726.4's hand-on, the venture walk
// through the Undercity, CR 704.5t, and restore points.

// initSettle resolves the stack, answering each prompt the Undercity
// asks: a trigger order in the order offered, a search with nothing, a
// scry keeping the order, a branch with the room `rooms` names next (in
// order). It stops at any other prompt, which the caller answers, and
// returns the rooms it was never asked for.
func InitSettle(t *testing.T, g *game.Game, rooms ...string) []string {
	t.Helper()
	for i := 0; i < 128; i++ {
		var open *game.PendingChoice
		for _, c := range g.PendingChoices {
			if c != nil {
				open = c
				break
			}
		}
		if open != nil {
			var err error
			switch open.Kind {
			case game.PendingChoiceTriggerOrder:
				err = g.ResolveTriggerOrder(open.ID, open.Chooser, append([]uuid.UUID(nil), open.TriggerOrderIDs...))
			case game.PendingChoiceSearchLibrary:
				err = g.ResolveSearchLibrary(open.ID, open.Chooser, nil)
			case game.PendingChoiceScry:
				err = g.ResolveScry(open.ID, open.Chooser, nil, append([]uuid.UUID(nil), open.ScryCards...))
			case game.PendingChoiceOptionPick:
				if len(rooms) == 0 || len(open.PickOptions) == 0 || open.PickOptions[0].Room == "" {
					return rooms
				}
				InitPickRoom(t, g, open.Chooser, rooms[0])
				rooms = rooms[1:]
			default:
				return rooms
			}
			if err != nil {
				t.Fatalf("answering %s: %v", open.Kind, err)
			}
			continue
		}
		if StackFullyEmpty(g) {
			return rooms
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority iter %d: %v", i, err)
		}
	}
	t.Fatal("the stack did not settle")
	return rooms
}

// initVenture has `player` take the initiative again (CR 726.5: they
// venture) and settles, choosing `rooms` at the branches.
func InitVenture(t *testing.T, g *game.Game, player uuid.UUID, rooms ...string) {
	t.Helper()
	InitTake(t, g, player)
	if left := InitSettle(t, g, rooms...); len(left) > 0 {
		t.Fatalf("rooms %v were never offered", left)
	}
}

// initCompleted reports `player`'s completed dungeons.
func InitCompleted(g *game.Game, player uuid.UUID) []string {
	for _, p := range g.Seats {
		if p.ID == player {
			return append([]string(nil), p.CompletedDungeons...)
		}
	}
	return nil
}
