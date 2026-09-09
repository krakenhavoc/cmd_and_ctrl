package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Krosan Grip — Instant {2}{G}:
//
//	"Split second. Destroy target artifact or enchantment."
//
// Sandbox simplification: SPLIT SECOND IS NOT IMPLEMENTED. In paper
// the point of this card over Naturalize is that nobody may respond
// while it's on the stack (CR 702.61) — here it is an ordinary
// instant and an opponent may hold priority and respond. Split
// second needs a cast-restriction layer the engine has no seam for
// yet, so it is deliberately deferred rather than faked; the
// destroy half is exactly right.
func init() {
	Register(Spec{
		OracleID: "3e39224c-72ce-4ecc-aa17-12c071ea1f3e",
		Name:     "Krosan Grip",
		Targets:  TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return DestroyTarget{Target: item.Targets[0].ID}.Apply(ctx)
		},
	})
}
