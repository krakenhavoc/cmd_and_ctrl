package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// block_tax_view_test.go — ADR 0221 (#5375): `turn.block_taxes` is the
// price on block staging, so the client shows it without re-deriving
// the rule.

const blockTaxViewOracle = "block-tax-view-test"

func TestBlockTaxesShowThePriceOfTheStagedBlocks(t *testing.T) {
	prev := game.CatalogBlockTaxes
	t.Cleanup(func() { game.CatalogBlockTaxes = prev })
	game.CatalogBlockTaxes = func(key string) []game.BlockTax {
		if key != blockTaxViewOracle {
			return nil
		}
		return []game.BlockTax{{
			Label:     "test block tax",
			AppliesTo: game.WhileSourceAttacking,
			ManaCost:  func(game.BlockTaxQuery) string { return "{1}" },
		}}
	}
	g := threeSeatsInDeclareAttackers(t)
	active := g.Seats[g.Turn.ActiveSeat]
	defIdx := (g.Turn.ActiveSeat + 1) % 3
	def := g.Seats[defIdx]
	angel := game.Card{
		InstanceID: uuid.New(), Name: "Angel", TypeLine: "Creature — Angel",
		Power: 3, Toughness: 5, OracleID: blockTaxViewOracle, Owner: active.ID, Controller: active.ID,
	}
	g.Battlefield.PushTop(angel)
	wall := game.Card{
		InstanceID: uuid.New(), Name: "Wall", TypeLine: "Creature — Wall",
		Power: 0, Toughness: 4, Owner: def.ID, Controller: def.ID,
	}
	g.Battlefield.PushTop(wall)
	if err := g.DeclareAttacker(angel.InstanceID, def.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	for i := 0; i < 5 && g.Turn.Step != game.StepDeclareBlockers; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}

	rows := ViewOfGame(g).Turn.BlockTaxes
	if len(rows) != 1 || rows[0].Seat != defIdx || rows[0].Cost != "" || rows[0].Each != "{1}" || len(rows[0].Labels) != 1 {
		t.Fatalf("before staging: %+v, want one row for the defender, nothing staged, {1} each", rows)
	}
	if err := g.DeclareBlocker(wall.InstanceID, angel.InstanceID); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	rows = ViewOfGame(g).Turn.BlockTaxes
	if len(rows) != 1 || rows[0].Cost != "{1}" {
		t.Fatalf("after staging one block: %+v, want cost {1}", rows)
	}
}
