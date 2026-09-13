package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Broken Bond — Sorcery {1}{G} (EDHREC rank 2234):
//
//	"Destroy target artifact or enchantment. You may put a land card
//	 from your hand onto the battlefield."
//
// Naturalize at sorcery speed with a free land drop attached. The
// destruction is the whole of what runs: a single target, re-checked
// at resolution.
//
// DECLARED SIMPLIFICATION — the Eureka Moment / Spelunking posture:
// "you may put a land card from your hand onto the battlefield" is
// not implemented. It is a pick-from-hand prompt with a
// hand-to-battlefield move, and neither exists (the Growth Spiral
// gap; the only hand picker the engine has is the discard modal).
// Shipping it as "put the first land" would be a choice the player
// never made, so the spell destroys its target and stops — weaker
// than printed, never stronger. It becomes whole the day a
// put-from-hand prompt lands.
func init() {
	Register(Spec{
		OracleID:     "858e12e9-3eaa-40cf-9e22-f9ccdfe485b3",
		Name:         "Broken Bond",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Only the artifact or enchantment is destroyed — it doesn't offer to put a land from your hand onto the battlefield."},
		Targets:      TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return DestroyTarget{Target: id}.Apply(ctx)
		},
	})
}
