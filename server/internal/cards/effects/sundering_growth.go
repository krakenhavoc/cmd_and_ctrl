package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sundering Growth — Instant {G/W}{G/W}:
//
//	"Destroy target artifact or enchantment, then populate. (Create a
//	 token that's a copy of a creature token you control.)"
//
// Disenchant with a populate behind it. The populate is not
// conditional on the destruction having happened: an indestructible
// target, or one regenerated, still lets you copy a token. It runs as
// the destroy's continuation so a prompt the destruction opens cannot
// be overtaken by the choice of what to copy.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a2a380d8-4df7-4357-862c-ed3fb795db6c",
		Name:         "Sundering Growth",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) > 0 && item.Targets[0].Kind == game.TargetCard {
				if err := (DestroyTarget{Target: item.Targets[0].ID}).Apply(ctx); err != nil {
					return err
				}
			}
			return Populate{}.Apply(ctx)
		},
	})
}
