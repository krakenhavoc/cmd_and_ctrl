package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Festergloom — Sorcery {2}{B} (Magic 2015, #5556):
//
//	"Nonblack creatures get -1/-1 until end of turn."
//
// Holy Light's body over nonblack creatures. The set is fixed as it
// resolves, read through the creatures' current colours, so a black
// creature is spared and one that arrives later this turn is not
// affected.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a653ac26-e9c9-4319-88f6-c83cc64d3bdb",
		Name:         "Festergloom",
		Completeness: effects.CompletenessFull,
		// ADR 0126 §6: a black creature is spared.
		Purpose: game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 1, Partial: true}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.NonBlack()), Power: -1, Toughness: -1, Label: "Festergloom — -1/-1"}.Apply(ctx)
		},
	})
}
