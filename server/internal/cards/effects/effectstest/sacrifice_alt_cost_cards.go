package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// handCardAtMain seeds a card into the active seat's hand at a main
// phase, for a cast that claims an alternative cost.
func HandCardAtMain(t *testing.T, g *game.Game, name, typeLine, oracle string) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Owner: me.ID, Controller: me.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	return id
}

func CastForAltCost(g *game.Game, id uuid.UUID, from, key string, pay []uuid.UUID, targets []game.TargetRef) error {
	return g.CastSpell(g.Seats[g.Turn.ActiveSeat].ID, id, game.CastSpellParams{
		FromZone:        from,
		AlternativeCost: key,
		AltCostIDs:      pay,
		Targets:         targets,
	})
}
