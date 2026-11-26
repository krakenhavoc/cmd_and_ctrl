package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chronic Flooding — Enchantment — Aura {1}{U} (Return to Ravnica, S82 OLD):
//
//	"Enchant land
//	 Whenever enchanted land becomes tapped, its controller mills three
//	 cards."
//
// Contaminated Ground's "becomes tapped" trigger (for mana or any other
// reason, CR 701.26a) with a mill in place of the life loss: the player who
// mills is the land's controller as the trigger resolves, read off the tapped
// land. A library holding fewer cards mills what it has (CR 701.17b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "474db5db-914e-474d-b0c2-7987d5986dc3",
		Name:         "Chronic Flooding",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantLand(),
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventTapCard, game.EventAttack},
			AppliesTo: effects.EnchantedPermanentBecameTapped,
			Key:       "Chronic Flooding — its controller mills three cards",
			Effect: func(g *game.Game, item *game.StackItem) error {
				if item.Trigger == nil {
					return nil
				}
				land, ok := g.LookupCardForEffect(item.Trigger.Event.CardID)
				if !ok {
					return nil
				}
				return effects.MillCards{Player: land.Controller, N: 3}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
