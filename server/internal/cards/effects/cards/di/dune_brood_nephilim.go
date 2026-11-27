package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dune-Brood Nephilim — Creature — Nephilim {B}{R}{G}{W}, 3/3 (Guildpact,
// S82 OLD batch 573, #5599):
//
//	"Whenever this creature deals combat damage to a player, create a 1/1
//	 colorless Sand creature token for each land you control."
//
// The lands are counted as the trigger resolves, under the controller's
// control and through the layered view (a land a layer 4 effect made, or lost,
// counts accordingly).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "634bd800-8caa-47ae-8b70-2c66baf9a355",
		Name:         "Dune-Brood Nephilim",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Dune-Brood Nephilim — create a 1/1 colorless Sand creature token for each land you control",
				effects.Do(effects.CreateTokensForEach{
					Template: effects.TokenCard("1/1 colorless Sand"),
					Name:     "land you control",
					Count: func(ctx *effects.Context) int {
						src, ok := ctx.Game.LookupCardForEffect(ctx.Source())
						if !ok {
							src = game.Card{Controller: ctx.Controller()}
						}
						src.Controller = ctx.Controller()
						return effects.CountPermanentsYouControl(ctx.Game, &src, func(c *game.Card) bool { return c.IsLand() })
					},
				})),
		},
	})
}
