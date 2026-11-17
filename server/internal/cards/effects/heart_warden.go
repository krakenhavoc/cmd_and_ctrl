package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Heart Warden — Creature — Elf Druid {1}{G}, 1/1 (Urza's Destiny,
// #5621):
//
//	"{T}: Add {G}.
//	 {2}, Sacrifice this creature: Draw a card."
//
// Llanowar Elves' mana ability, which waits out summoning sickness
// (CR 302.6) and doesn't use the stack (CR 605.1a), plus Illvoi
// Galeblade's sacrifice-to-draw row. The draw row has no {T}, so it can
// be activated the turn the Elf arrives, and an Elf that has tapped for
// {G} can still be sacrificed for a card.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "babcc551-39bb-4b1c-92a0-e60b236a59b2",
		Name:         "Heart Warden",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "Add {G}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{2}, Sacrifice this creature: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{2}"), SacrificeThis()),
			Effect:  ebDrawACard,
		}},
	})
}
