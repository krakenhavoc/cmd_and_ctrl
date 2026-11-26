package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lagonna-Band Elder — Creature — Centaur Advisor {2}{W}, 3/2 (Theros, S82 OLD b391, #5561):
//
//	"When this creature enters, if you control an enchantment, you gain 3
//	 life."
//
// Scholar of Stars' intervening if (CR 603.4), checked as the trigger fires
// and again as it resolves, over the controller's enchantments.
//
// No simplification.
func init() {
	controlsAnEnchantment := func(g *game.Game, source *game.Card) bool {
		return effects.CountPermanentsYouControl(g, source, func(c *game.Card) bool { return c.IsEnchantment() }) > 0
	}
	effects.Register(effects.Spec{
		OracleID:     "60e3ccf2-63f5-4435-a752-1d82fefb55d7",
		Name:         "Lagonna-Band Elder",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondBoard{
				Name: "you control an enchantment",
				Test: func(ctx *effects.Context) bool {
					src, ok := ctx.Game.LookupCardForEffect(ctx.Source())
					return ok && controlsAnEnchantment(ctx.Game, &src)
				},
			}, effects.On(game.EventETB, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return effects.Self(ev, source, lki, g) && controlsAnEnchantment(g, source)
			}, "Lagonna-Band Elder — you gain 3 life", effects.Do(effects.GainLife{Amount: 3}))),
		},
	})
}
