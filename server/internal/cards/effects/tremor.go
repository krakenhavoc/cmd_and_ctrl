package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID: "4281a153-665b-44cf-bcff-9cf88c14c2b5",
		Name:     "Tremor",
		// ADR 0126 §6: a creature with flying is spared.
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1, Partial: true}},
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return damageEachMatching(ctx, And(Creature(), WithoutKeyword("flying")), 1)
		},
	})
}
