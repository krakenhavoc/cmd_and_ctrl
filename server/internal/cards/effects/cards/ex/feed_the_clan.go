package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Feed the Clan — Khans of Tarkir (S82 OLD b347, #5554):
//
//	Instant {1}{G}: You gain 5 life. Ferocious — You gain 10 life instead if you
//	control a creature with power 4 or greater.
//
// Ferocious is read through YouControlPowerFourOrGreater (current power, layered).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f70f1f2e-7277-4bc9-b0bb-b44a1da2c62f",
		Name:         "Feed the Clan",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			n := 5
			if effects.FerociousYouControlPowerFourOrGreater(ctx) {
				n = 10
			}
			return effects.GainLife{Amount: n}.Apply(ctx)
		},
	})
}
