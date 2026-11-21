package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pyromancy — Enchantment {2}{R}{R}:
//
//	"{3}, Discard a card at random: This enchantment deals damage to any
//	 target equal to the mana value of the discarded card."
//
// ADR 0109 §7, owner decision 3: "Discard a card at random" is a cost
// the activator chooses nothing for (CR 701.9b). The engine draws the
// card from the hand the rest of the cost leaves, paid after every other
// cost (CR 601.2h), and an empty hand can't pay it (CR 118.3).
// The amount is the discarded card's mana value, read off the payment
// record (Context.DiscardedManaValue, ADR 0109 §8, CR 400.7j) as the
// ability resolves. A land, or a card with no mana cost, deals 0.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "71a176ce-cbe7-4cec-a2ca-58abf59ad964",
		Name:         "Pyromancy",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, Discard a card at random: This enchantment deals damage to any target equal to the mana value of the discarded card.",
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.DiscardAtRandom(1, "a card at random")),
			Targets: effects.TargetAny(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.SourceDealsDamageToEachLegalTarget(effects.NewContext(g, item).DiscardedManaValue())(g, item)
			},
		}},
	})
}
