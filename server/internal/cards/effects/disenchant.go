package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Disenchant — Instant {1}{W} (EDHREC rank 1103):
//
//	"Destroy target artifact or enchantment."
//
// The original. Naturalize in white: the same target clause and the
// same single primitive.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a7e97fa9-4b72-4548-b854-5be5f18a6f1a",
		Name:         "Disenchant",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return DestroyTarget{Target: item.Targets[0].ID}.Apply(ctx)
		},
	})
}
