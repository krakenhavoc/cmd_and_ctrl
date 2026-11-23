package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// until_end_of_turn_test.go is the card-level half of S32: Giant
// Growth and Overrun driven through a real cast, with the layer
// engine reading the result. The registry's own mechanics (undo,
// layer ordering, the end-step expiry rule) are pinned in
// server/internal/game/turn_scoped_statics_test.go.

const (
	GiantGrowthOracle = "5748ebf1-24e3-499d-ab7c-c2cebd462a24"
	OverrunOracle     = "204f9afe-c20b-4933-b5cd-aa572784762a"
)

// advanceToNextSeatsTurn walks the cursor past this seat's cleanup
// step. Turn.Round counts rounds, so the active seat is the
// reliable "the turn ended" marker.
func AdvanceToNextSeatsTurn(t *testing.T, g *game.Game) {
	t.Helper()
	start := g.Turn.ActiveSeat
	for i := 0; i < 30; i++ {
		if g.Turn.ActiveSeat != start {
			return
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	t.Fatalf("cursor never left seat %d's turn", start)
}

// hasEffectiveKeyword forces a recompute and reports whether the
// keyword is in the card's post-layer abilities — the same read
// game.HasKeyword does from the combat code.
func HasEffectiveKeyword(t *testing.T, g *game.Game, cardID uuid.UUID, kw string) bool {
	t.Helper()
	found := false
	present := false
	g.ReadSnapshot(func() {
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			if c.InstanceID != cardID {
				continue
			}
			found = true
			present = game.HasKeyword(c, kw)
			return
		}
	})
	if !found {
		t.Fatalf("card %s not on battlefield", cardID)
	}
	return present
}
