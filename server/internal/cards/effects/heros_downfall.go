package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hero's Downfall — Instant {1}{B}{B} (EDHREC rank 1977):
//
//	"Destroy target creature or planeswalker."
//
// The clean black answer to either kind of threat. One target
// clause over both types; the single-target verb honours
// indestructible (S25).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "03df6a57-37c9-46d3-83b3-4a6240100714",
		Name:         "Hero's Downfall",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return DestroyTarget{Target: item.Targets[0].ID}.Apply(ctx)
		},
	})
}
