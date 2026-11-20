package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "a0802918-e18a-4716-befb-96ae6667f2f3",
		Name:         "Infernal Idol",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{B}",
			Label:    "Add {B}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}{B}, {T}, Sacrifice this artifact: You draw two cards and lose 2 life.",
			Cost:    effects.Plus(effects.ManaCost("{1}{B}{B}"), effects.TapCost(), effects.SacrificeThis()),
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 2},
			Effect: effects.Do(
				effects.DrawCards{N: 2},
				effects.YouLoseLife(2),
			),
		}},
	})
}
