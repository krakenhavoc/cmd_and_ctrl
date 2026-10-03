package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stormscale Anarch — Creature — Lizard Shaman {2}{R}{R}, 2/2:
//
//	"{2}{R}, Discard a card at random: This creature deals 2 damage to
//	 any target. If the discarded card was multicolored, this creature
//	 deals 4 damage instead."
//
// ADR 0109 §7, owner decision 3: "Discard a card at random" is a cost
// the activator chooses nothing for (CR 701.9b). The engine draws the
// card from the hand the rest of the cost leaves, paid after every other
// cost (CR 601.2h), and an empty hand can't pay it (CR 118.3).
// "The discarded card" is read off the payment record
// (Context.DiscardedCard, ADR 0109 §8, CR 400.7j) as the ability
// resolves: two or more colours is multicolored (CR 105.2b), and the
// damage is 4 instead of 2.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "819a7235-b61a-49d7-a0ad-3d8bbf83091e",
		Name:         "Stormscale Anarch",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}{R}, Discard a card at random: This creature deals 2 damage to any target. If the discarded card was multicolored, this creature deals 4 damage instead.",
			Cost:    Plus(ManaCost("{2}{R}"), DiscardAtRandom(1, "a card at random")),
			Targets: TargetAny(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				amount := 2
				if card, ok := NewContext(g, item).DiscardedCard(); ok && len(card.EffectiveColors()) >= 2 {
					amount = 4
				}
				return sourceDealsDamageToEachLegalTarget(amount)(g, item)
			},
		}},
	})
}
