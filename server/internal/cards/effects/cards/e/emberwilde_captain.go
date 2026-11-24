package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Emberwilde Captain — Creature — Djinn Pirate {3}{R}, 4/2 (Commander
// Legends):
//
//	"When this creature enters, you become the monarch.
//	 Whenever an opponent attacks you while you're the monarch, this
//	 creature deals damage to that player equal to the number of cards in
//	 their hand."
//
// The entry trigger is the monarch grant. The second is Lulu's "an opponent
// attacks you": one occurrence per attacking player however many creatures
// they send at the controller (OncePerBatchPerAttacker), and only while the
// controller is the monarch as the attack is declared (a condition of the
// trigger, not an intervening "if"). The attacker rides the item, so the
// damage goes to the right player, for the cards in their hand as the
// trigger resolves.
//
// No simplification.
func init() {
	const label = "Emberwilde Captain — it deals damage to that player equal to the number of cards in their hand"
	effects.Register(effects.Spec{
		OracleID:     "4f4ec0fa-fcaf-4dc2-9c67-ee7881f5f77d",
		Name:         "Emberwilde Captain",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersYouBecomeTheMonarch("Emberwilde Captain"),
			effects.OncePerBatchPerAttacker(game.TriggeredAbility{
				Watches: []game.EventKind{game.EventAttack},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return effects.B17OpponentsCreatureAttackedYou(ev, source, g) && effects.YoureTheMonarch(g, source.Controller)
				},
				Key: label,
				Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					item := game.NewTriggeredItem(source, label)
					item.Params.Player = ev.Actor
					return item
				},
				Effect: func(g *game.Game, item *game.StackItem) error {
					victim := g.PlayerByIDForEffect(item.Params.Player)
					if victim == nil || victim.Eliminated || victim.Hand == nil || victim.Hand.Size() == 0 {
						return nil
					}
					return effects.DealDamage{Source: item.SourceCardID, Target: victim.ID, Amount: victim.Hand.Size()}.Apply(effects.NewContext(g, item))
				},
			}),
		},
	})
}
