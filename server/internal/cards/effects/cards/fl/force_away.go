package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Force Away — Khans of Tarkir (S82 OLD b347, #5554):
//
//	Instant {1}{U}: Return target creature to its owner's hand. Ferocious — If you
//	control a creature with power 4 or greater, you may draw a card. If you do,
//	discard a card.
//
// Ferocious is read through YouControlPowerFourOrGreater (current power, layered). The loot is the may-draw-then-discard of Marauding Looter.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4770ec88-bc7b-457e-8f67-ae46e2cf4caa",
		Name:         "Force Away",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.BounceTheTarget(item, ctx); err != nil {
				return err
			}
			if !effects.FerociousYouControlPowerFourOrGreater(ctx) {
				return nil
			}
			return effects.OldB223MayLoot("Force Away — ferocious: draw a card, then discard a card?")(ctx.Game, item)
		},
	})
}
