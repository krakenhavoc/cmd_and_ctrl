package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jennika's Technique — Instant {2}{R} (TMT, #4457):
//
//	"Sneak {R} (You may cast this spell for {R} if you also return an
//	 unblocked attacker you control to hand during the declare blockers
//	 step.)
//	 Jennika's Technique deals 2 damage to each creature."
//
// Sneak is Sneak(cost) (CR 702.190); the body is Pyroclasm's. The
// creature returned to pay the sneak cost is in its owner's hand by
// then, so it is spared.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "90b1226f-4c90-47fd-98f3-ad4edd3b3518",
		Name:             "Jennika's Technique",
		Purpose:          game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 2}},
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Sneak("{R}")},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return damageEachMatching(ctx, Creature(), 2)
		},
	})
}
