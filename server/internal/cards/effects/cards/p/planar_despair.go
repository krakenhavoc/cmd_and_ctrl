package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Planar Despair — Sorcery {3}{B}{B} (Apocalypse, S82 OLD, #5614):
//
//	"Domain — All creatures get -1/-1 until end of turn for each basic land
//	 type among lands you control."
//
// Mutilate's sweep with the shared domain count (CR 305.6): counted once as the
// spell resolves, and only the creatures on the battlefield then are affected
// (CR 611.2c). With no basic land type it does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID: "1ba36042-8902-4abb-a4ae-a8d26d81a0de",
		Name:     "Planar Despair",
		// The amount is counted at resolution, like an X (AmountIsX).
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, AmountIsX: true}},
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			n := effects.DomainCount(ctx.Game, ctx.Controller())
			if n <= 0 {
				return nil
			}
			return effects.BoostUntilEOT{Match: effects.Creature(), Power: -n, Toughness: -n,
				Label: "Planar Despair — -1/-1 for each basic land type"}.Apply(ctx)
		},
	})
}
