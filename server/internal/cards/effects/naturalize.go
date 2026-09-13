package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Naturalize — Instant {1}{G}:
//
//	"Destroy target artifact or enchantment."
//
// The baseline the green disenchants are measured against — Krosan
// Grip costs one more for split second, Nature's Claim one less for
// giving the controller 4 life.
func init() {
	Register(Spec{
		OracleID:     "bdb3ca68-ec1f-4e16-81cc-d23f8f52c728",
		Name:         "Naturalize",
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
