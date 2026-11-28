package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ghitu Fire — Sorcery {X}{R} (Invasion, S82 #5616):
//
//	"You may cast this spell as though it had flash if you pay {2} more to cast
//	 it. (You may cast it any time you could cast an instant.)
//	 Ghitu Fire deals X damage to any target."
//
// Blaze's X damage with Rout's optional {2} that opens the instant-speed window
// (FlashForMore, CR 601.3c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "c15ae9d8-b423-4628-ad88-3f44515e0c77",
		Name:          "Ghitu Fire",
		Completeness:  effects.CompletenessFull,
		XMatters:      true,
		OptionalCosts: []game.AdditionalCost{effects.FlashForMore("{2}")},
		Targets:       effects.TargetAny(),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			return effects.DealDamage{Source: ctx.Source(), Target: item.Targets[0].ID, Amount: ctx.X()}.Apply(ctx)
		},
	})
}
