package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "73144877-1d37-4c7a-8493-3cef191b9211",
		Name:         "Omni-Cheese Pizza",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.SamiDrawOnETB("Omni-Cheese Pizza")},
		ManaAbilities: []effects.ManaAbility{
			effects.SamiAnyColorMana(effects.ManaAbilityCost{Tap: true, Sacrifice: true, Mana: "{1}"},
				"{1}, {T}, Sacrifice this artifact: Add one mana of any color"),
		},
		Activated: []effects.ActivatedAbility{effects.SamiFoodLifeAbility()},
	})
}
