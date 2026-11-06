package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Molt Tender — Creature — Insect Druid {G}, 1/1 (Aetherdrift, #4464):
//
//	"{T}: Mill a card. (Put the top card of your library into your
//	 graveyard.)
//	 {T}, Exile a card from your graveyard: Add one mana of any color."
//
// The mill is an ordinary activated ability (CR 701.17a) that uses the
// stack. The second is a mana ability (CR 605.1a) whose exile is a cost
// the activator picks, Rubble Rouser's ExileCardsFromGraveyard component:
// with an empty graveyard it cannot be activated. Both tap the Insect, so
// summoning sickness stops each and only one can be used per untap.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9a39ca64-e1a9-42ef-8efd-6aa3edde3d93",
		Name:         "Molt Tender",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: Mill a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    TapCost(),
			Effect:  Do(MillCards{N: 1}),
		}},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true, ExileCards: ExileCardsFromGraveyard(1, "a card", nil)},
			Produced: "{W|U|B|R|G}",
			Label:    "{T}, Exile a card from your graveyard: Add one mana of any color.",
		}},
	})
}
