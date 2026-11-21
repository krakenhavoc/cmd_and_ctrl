package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tremor — Sorcery {R} (Visions, #5631):
//
//	"Tremor deals 1 damage to each creature without flying."
//
// Seismic Rupture's sweep at 1: the creatures without flying are read
// post-layer as it resolves, so one granted flying is spared, and no
// player is dealt damage.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID: "4281a153-665b-44cf-bcff-9cf88c14c2b5",
		Name:     "Tremor",
		// ADR 0126 §6: a creature with flying is spared.
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1, Partial: true}},
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DamageEachMatching(ctx, effects.And(effects.Creature(), effects.WithoutKeyword("flying")), 1)
		},
	})
}
