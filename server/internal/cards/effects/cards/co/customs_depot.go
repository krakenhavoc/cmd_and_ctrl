package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Customs Depot — Enchantment {1}{U} (Mercadian Masques, S82 OLD, #5619):
//
//	"Whenever you cast a creature spell, you may pay {1}. If you do, draw a
//	 card, then discard a card."
//
// Mind's Eye's may-pay for {1}, with a loot on the paid branch: the draw
// resolves first and the discard prompt opens after it, so the drawn card may be
// discarded. A payment the controller can't fund is a decline.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "de64eab5-9d7a-4408-8358-47d30a14ce40",
		Name:         "Customs Depot",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Creature(), "Customs Depot — you may pay {1}; if you do, draw a card, then discard a card",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{1}",
						Question: "Customs Depot — pay {1} to draw a card, then discard a card?",
						Then:     customsDepotLoot,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// customsDepotLoot is the paid branch. An on-disk key: never renamed, never
// reused.
var customsDepotLoot = effects.PayUnlessBranches("pay-unless/customs-depot-loot",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.DrawThenDiscard(ctx.Game, ctx.Controller(), ctx.Source(), 1, 1, "")
	}, nil)
