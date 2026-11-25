package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lamplighter of Selhoff — Creature — Zombie Horror {4}{U}, 3/5 (Shadows
// over Innistrad, #5543):
//
//	"When this creature enters, if you control another Zombie, you may draw
//	 a card. If you do, discard a card."
//
// An intervening if (CR 603.4): another Zombie (not this creature) under the
// controller's control as it enters and again as the trigger resolves. The
// loot is The Celestus's "may draw, if you do discard": declining draws and
// discards nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a88c5540-a242-49bc-a725-39b9fda085c3",
		Name:         "Lamplighter of Selhoff",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1, Discards: 1},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondBoard{
				Name: "you control another Zombie",
				Test: func(ctx *effects.Context) bool {
					for _, c := range ctx.Game.BattlefieldCardsForEffect() {
						if c.Controller == ctx.Controller() && c.InstanceID != ctx.Source() && c.HasSubtype(ctx.TextSubtype("Zombie")) {
							return true
						}
					}
					return false
				},
			}, effects.WhenThisEnters("Lamplighter of Selhoff — you may draw a card; if you do, discard a card",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayChoice{
						Question: "Lamplighter of Selhoff — draw a card, then discard a card?",
						Draws:    1,
						YesLabel: "Draw",
						NoLabel:  "Decline",
						Then:     effects.CelestusAnswer,
					}.Apply(effects.NewContext(g, item))
				})),
		},
	})
}
