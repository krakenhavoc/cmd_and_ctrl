package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chain Reaction — Sorcery {2}{R}{R}:
//
//	"Chain Reaction deals X damage to each creature, where X is the
//	 number of creatures on the battlefield."
//
// X is counted once, as the spell resolves, before any damage is
// dealt, and every creature takes that same X (a creature that dies
// to it does so at the state-based sweep afterwards, so the count
// does not shrink mid-resolution). It counts every creature on the
// battlefield, the caster's included. Damage, not destruction:
// indestructible and prevention shields work as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "086b2564-9114-4ba2-94fd-b490f98f38a7",
		Name:         "Chain Reaction",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return damageEachMatching(ctx, Creature(), len(MatchingBattlefield(ctx, Creature())))
		},
	})
}
