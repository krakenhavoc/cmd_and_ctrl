package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Omni-Cheese Pizza — {2} Artifact — Food (Teenage Mutant Ninja Turtles,
// #4457):
//
//	"When this artifact enters, draw a card.
//	 {1}, {T}, Sacrifice this artifact: Add one mana of any color.
//	 {2}, {T}, Sacrifice this artifact: You gain 3 life."
//
// Golden Egg's text under another name: the mana half is a mana ability
// (CR 605.1a, no stack), the life half is the Food ability on the stack.
// The Food subtype comes off the type line, so Food payoffs see it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "73144877-1d37-4c7a-8493-3cef191b9211",
		Name:         "Omni-Cheese Pizza",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{samiDrawOnETB("Omni-Cheese Pizza")},
		ManaAbilities: []ManaAbility{
			samiAnyColorMana(ManaAbilityCost{Tap: true, Sacrifice: true, Mana: "{1}"},
				"{1}, {T}, Sacrifice this artifact: Add one mana of any color"),
		},
		Activated: []ActivatedAbility{samiFoodLifeAbility()},
	})
}
