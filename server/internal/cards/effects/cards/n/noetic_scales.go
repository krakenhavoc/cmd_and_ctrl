package cardsn

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Noetic Scales — Artifact {4} (Urza's Saga, S82 OLD batch 732, #5624):
//
//	"At the beginning of each player's upkeep, return to its owner's hand each
//	 creature that player controls with power greater than the number of cards
//	 in their hand."
//
// "That player" is the one whose upkeep it is. The creatures are found as the
// trigger resolves, their power read through the layers (counters included,
// CR 208.1) against the player's hand size then, and they return together to
// their owners' hands (CR 700.4); tokens cease to exist. Nothing is chosen.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "aebddadd-f05b-4ad5-bda4-603a134f68dc",
		Name:         "Noetic Scales",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtEachUpkeep("Noetic Scales — return each creature that player controls with power greater than the number of cards in their hand to its owner's hand",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if item.Trigger == nil {
						return nil
					}
					player := item.Trigger.Event.Actor
					p := g.PlayerByIDForEffect(player)
					if p == nil || p.Hand == nil {
						return nil
					}
					hand := len(p.Hand.Cards)
					var ids []uuid.UUID
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller == player && c.IsCreature() && c.CurrentPower() > hand {
							ids = append(ids, c.InstanceID)
						}
					}
					return effects.ReturnAllToHand{Cards: ids}.Apply(ctx)
				}),
		},
	})
}
