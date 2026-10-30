package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fogwell's Gym — Land (Marvel Super Heroes Commander, #4455):
//
//	"{T}: Add {R}. This land deals 1 damage to you.
//	 {2}{R}, {T}, Discard a card: Draw a card."
//
// Barbarian Ring's painful red mana and a looting-style activation whose
// costs are all paid at announcement (CR 602.2b); the draw cannot be
// responded to before the discard happened.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "850bb6f7-48d3-4d65-9220-b0bec5ee6b64",
		Name:         "Fogwell's Gym",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:      ManaAbilityCost{Tap: true},
			Produced:  "{R}",
			Label:     "Add {R}. This land deals 1 damage to you.",
			PainToYou: 1,
		}},
		Activated: []ActivatedAbility{{
			Label:   "{2}{R}, {T}, Discard a card: Draw a card.",
			Cost:    Plus(ManaCost("{2}{R}"), TapCost(), DiscardACard()),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect:  Do(DrawCards{N: 1}),
		}},
	})
}
