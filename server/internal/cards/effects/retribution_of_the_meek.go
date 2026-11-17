package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Retribution of the Meek — Sorcery {2}{W} (Visions, #5631):
//
//	"Destroy all creatures with power 4 or greater. They can't be
//	 regenerated."
//
// Power is read post-layer as it resolves, so a creature pumped to 4 is
// destroyed and one shrunk below it is spared. The rider is CR 701.19c.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID: "61f2083b-3018-43a4-8852-8e6d2a97d5d6",
		Name:     "Retribution of the Meek",
		// ADR 0126 §6: a creature with power 3 or less is spared.
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}},
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DestroyAllMatching{Match: And(Creature(), PowerGE(4)), CantBeRegenerated: true}.Apply(ctx)
		},
	})
}
