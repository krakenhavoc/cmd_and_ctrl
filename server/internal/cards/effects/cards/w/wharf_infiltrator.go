package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wharf Infiltrator — Creature — Human Horror {1}{U}, 1/1 (Commander Legends:
// Battle for Baldur's Gate, #5948):
//
//	"Skulk (This creature can't be blocked by creatures with greater power.)
//	 Whenever this creature deals combat damage to a player, you may draw a
//	 card. If you do, discard a card.
//	 Whenever you discard a creature card, you may pay {2}. If you do, create a
//	 3/2 colorless Eldrazi Horror creature token."
//
// Skulk is the engine keyword. The loot is Rook Turret's (a "may" whose yes
// draws, then discards); the discard trigger is Surly Badgersaur's creature
// discard watcher (it fires for the Infiltrator's own loot too) with Kavaron
// Harrier's pay-{2} choice leading to the token.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ac88975c-fb2c-441a-bdd4-e7de6ba865e7",
		Name:            "Wharf Infiltrator",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"skulk"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Wharf Infiltrator — you may draw a card; if you do, discard a card",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayChoice{
						Question: "Wharf Infiltrator — draw a card, then discard a card?",
						Draws:    1,
						YesLabel: "Draw",
						NoLabel:  "Decline",
						Then:     effects.CelestusAnswer,
					}.Apply(effects.NewContext(g, item))
				}),
			effects.On(game.EventDiscardCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B30YouDiscardedCardWhere(ev, source, g, func(c game.Card) bool { return c.IsCreature() })
			}, "Wharf Infiltrator — you may pay {2}; if you do, create a 3/2 colorless Eldrazi Horror creature token",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{2}",
						Question: "Wharf Infiltrator — pay {2} to create a 3/2 Eldrazi Horror?",
						Then:     wharfInfiltratorPaid,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

var wharfInfiltratorPaid = effects.PayUnlessBranches("pay-unless/wharf-infiltrator-eldrazi-horror",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.CreateToken{Controller: ctx.Controller(), Template: effects.TokenCard("3/2 colorless Eldrazi Horror"), N: 1}.Apply(ctx)
	}, nil)
