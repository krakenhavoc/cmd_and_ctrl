package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fracture — Instant {W}{B} (EDHREC rank 1256):
//
//	"Destroy target artifact, enchantment, or planeswalker."
//
// Mortify's shape with a three-way target clause; Despark's
// colours. The target predicate is the whole card.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f21d0319-0509-4ac1-b6e3-10955a26fd7a",
		Name:         "Fracture",
		Completeness: CompletenessFull,
		Targets: TargetPermanent("target artifact, enchantment, or planeswalker",
			Or(Artifact(), Enchantment(), Planeswalker())),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return DestroyTarget{Target: item.Targets[0].ID}.Apply(ctx)
		},
	})
}
