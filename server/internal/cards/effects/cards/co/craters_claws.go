package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Crater's Claws — Khans of Tarkir (S82 OLD b347, #5554):
//
//	Sorcery {X}{R}: Crater's Claws deals X damage to any target. Ferocious —
//	Crater's Claws deals X plus 2 damage instead if you control a creature with
//	power 4 or greater.
//
// Ferocious is read through YouControlPowerFourOrGreater (current power, layered).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b8cf8614-14f4-4c20-9697-2e5ea34e0039",
		Name:         "Crater's Claws",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Targets:      effects.TargetAny(),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			n := ctx.X()
			if effects.FerociousYouControlPowerFourOrGreater(ctx) {
				n += 2
			}
			return effects.DealDamage{Source: ctx.Source(), Target: item.Targets[0].ID, Amount: n}.Apply(ctx)
		},
	})
}
