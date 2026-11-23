package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func RfCardTarget(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

// rfCastModal casts a modal instant from the active seat's hand with
// the given modes and targets.
func RfCastModal(t *testing.T, g *game.Game, name, oracle string, modes []int, targets []game.TargetRef) error {
	t.Helper()
	ToMainForCost(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	id := HandCardFull(me, name, "Instant", "", oracle, nil)
	return g.CastSpell(me.ID, id, game.CastSpellParams{Modes: modes, Targets: targets})
}

// rfPT is the creature's power and toughness INCLUDING counters, which
// effectivePower / effectiveToughness (layer output only) leave out.
func RfPT(t *testing.T, g *game.Game, id uuid.UUID) (int, int) {
	t.Helper()
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		t.Fatalf("card %s not found", id)
	}
	return c.CurrentPower(), c.CurrentToughness()
}
