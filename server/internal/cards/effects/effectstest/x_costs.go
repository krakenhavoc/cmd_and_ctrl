package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// x_costs_test.go — S20 sub-PR 3: X spells read the announced X at
// resolution; the cost gate charges it.

const (
	BlazeOracle          = "0596920f-9946-42f4-a03b-24aab67f9f1b"
	ExsanguinateOracle   = "8164b1e8-3350-465e-8a17-75f57d326344"
	StrokeOfGeniusOracle = "0cc6d683-366f-4ae4-be60-20ad9621fdaf"
)

// castXSpell seeds an X card into the active seat's hand and casts
// it with the given X + targets (permissive cost mode — the S15
// gate warns rather than blocks without a pool).
func CastXSpell(t *testing.T, g *game.Game, name, typeLine, oracle, manaCost string, x int, targets []game.TargetRef) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, ManaCost: manaCost,
		Owner: active.ID, Controller: active.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{Targets: targets, XValue: x}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}
