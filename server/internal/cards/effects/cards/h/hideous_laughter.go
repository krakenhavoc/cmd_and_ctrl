package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hideous Laughter — Instant — Arcane {2}{B}{B}:
//
//	"All creatures get -2/-2 until end of turn.
//	Splice onto Arcane {3}{B}{B}"
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e9022d1c-2f5c-4a1f-a44e-0e4f9651e729",
		Name:         "Hideous Laughter",
		Completeness: effects.CompletenessFull,
		Splice:       effects.SpliceOntoArcane("{3}{B}{B}"),
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 2}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.BoostUntilEOT{Match: effects.Creature(), Power: -2, Toughness: -2, Label: "Hideous Laughter — -2/-2"}.Apply(ctx)
		},
	})
}
