package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mirror Entity — Creature — Shapeshifter {2}{W}, 1/1:
//
//	"Changeling (This card is every creature type.)
//	 {X}: Until end of turn, creatures you control have base power
//	 and toughness X/X and gain all creature types."
//
// Changeling needs no catalog entry (deck import stamps it, S26).
// The activated ability is an {X} cost (`Soothsaying`'s shape) over
// two CR 611.2c effects folded into one `untilEndOfTurn` call: a base
// P/T SET (`SetBasePowerMod` / `SetBaseToughnessMod` — Layer 7b, not
// the Layer 7c pump `BoostUntilEOT` uses) and an all-creature-types
// grant (`AllCreatureTypesMod`), both snapshotted once against
// "creatures you control" at activation. X=0 is legal and turns the
// controller's board into a wall of 0/0s that die to the state-based
// action a moment later — the printed card, not a bug.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "17e905ca-c0bd-473d-95a7-e180ba5fea43",
		Name:         "Mirror Entity",
		Completeness: CompletenessFull,
		XMatters:     true,
		Activated: []ActivatedAbility{{
			Label: "{X}: Until end of turn, creatures you control have base power and toughness X/X and gain all creature types.",
			Cost:  ManaCost("{X}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				x := ctx.X()
				return untilEndOfTurn(ctx, uuid.Nil, And(Creature(), YouControl()),
					"Mirror Entity — base power and toughness X/X and all creature types",
					game.SetBasePowerMod(x), game.SetBaseToughnessMod(x), game.AllCreatureTypesMod())
			},
		}},
	})
}
