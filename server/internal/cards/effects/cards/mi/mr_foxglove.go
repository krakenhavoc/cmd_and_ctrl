package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mr. Foxglove — Legendary Creature — Fox Rogue {2}{G}{W}{U}, 3/5
// (Bloomburrow Commander, #5368):
//
//	"Lifelink
//	 Whenever Mr. Foxglove attacks, draw cards equal to the number of
//	 cards in defending player's hand minus the number of cards in your
//	 hand. If you didn't draw cards this way, you may put a creature card
//	 from your hand onto the battlefield."
//
// Both hand sizes are read as the trigger resolves. The defending player is
// the player attacked, or the controller of the planeswalker or battle
// attacked. A difference of zero or less draws nothing, and then the
// controller may put a creature card from their hand onto the battlefield.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c2008ba9-00df-4607-ba0c-189af52033eb",
		Name:            "Mr. Foxglove",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Mr. Foxglove — draw cards equal to the defending player's hand minus yours; if you didn't, you may put a creature card from your hand onto the battlefield",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					ctx := effects.NewContext(g, item)
					theirs, mine := 0, 0
					if d := g.PlayerByIDForEffect(effects.DefendingPlayerOfAttack(g, item.Trigger.Event)); d != nil {
						theirs = d.Hand.Size()
					}
					if p := g.PlayerByIDForEffect(item.Controller); p != nil {
						mine = p.Hand.Size()
					}
					if n := theirs - mine; n > 0 {
						return effects.DrawCards{N: n}.Apply(ctx)
					}
					return effects.PutFromHandOntoBattlefield{
						Match:    effects.Creature(),
						Optional: true,
						Label:    "Mr. Foxglove — you may put a creature card from your hand onto the battlefield",
					}.Apply(ctx)
				}),
		},
	})
}
