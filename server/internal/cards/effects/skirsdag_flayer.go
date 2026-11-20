package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Skirsdag Flayer — Creature — Human Cleric {1}{B}, 1/1 (Dark Ascension, #5571):
//
//	"{3}{B}, {T}, Sacrifice a Human: Destroy target creature."
//
// The Human is sacrificed as part of the cost (CR 602.2b); the Flayer is a
// Human itself and may pay with its own body, as in paper.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "38ea058d-f8cd-4bd9-9142-8f0d3d56d571",
		Name:         "Skirsdag Flayer",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{3}{B}, {T}, Sacrifice a Human: Destroy target creature.",
			Purpose: game.Purpose{Answers: game.AnswerRemove | game.AnswerSacOutlet},
			Cost:    game.AbilityCost{Mana: "{3}{B}", Tap: true, SacrificeOther: sacrificeSpec("a Human", HasSubtype("Human"))},
			Targets: TargetCreature("target creature"),
			Effect:  destroyFirstLegalCardTarget,
		}},
	})
}
