package effectstest

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// pr8CastX casts a catalogued spell from the active player's hand with X.
func Pr8CastX(t *testing.T, g *game.Game, name, typeLine, oracle string, x int, targets []game.TargetRef) {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	c := game.NewCard(name, active.ID)
	c.TypeLine, c.OracleID = typeLine, oracle
	active.Hand.PushTop(c)
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	if err := g.CastSpell(active.ID, c.InstanceID, game.CastSpellParams{Targets: targets, XValue: x}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	PassPriorityAroundTable(t, g)
}
