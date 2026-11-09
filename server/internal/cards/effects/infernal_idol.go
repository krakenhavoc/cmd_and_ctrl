package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Infernal Idol — Artifact {3}:
//
//	"{T}: Add {B}.
//	 {1}{B}{B}, {T}, Sacrifice this artifact: You draw two cards and lose
//	 2 life."
//
// A mana ability and Chromatic Sphere's sacrifice-for-cards shape. Both
// halves are one resolution, the draw first (CR 608.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a0802918-e18a-4716-befb-96ae6667f2f3",
		Name:         "Infernal Idol",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{B}",
			Label:    "Add {B}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{1}{B}{B}, {T}, Sacrifice this artifact: You draw two cards and lose 2 life.",
			Cost:    Plus(ManaCost("{1}{B}{B}"), TapCost(), SacrificeThis()),
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 2},
			Effect: Do(
				DrawCards{N: 2},
				youLoseLife(2),
			),
		}},
	})
}
