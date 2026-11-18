package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Executioner's Capsule — Artifact {B}:
//
//	"{1}{B}, {T}, Sacrifice this artifact: Destroy target nonblack
//	 creature."
//
// Goblin Firebomb's pay-everything-at-announce destroy row with Doom
// Blade's nonblack target, re-checked at resolution (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c9cd266c-7ecf-4beb-b9da-69b88f33abd3",
		Name:         "Executioner's Capsule",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}{B}, {T}, Sacrifice this artifact: Destroy target nonblack creature.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    game.AbilityCost{Tap: true, SacrificeSelf: true, Mana: "{1}{B}"},
			Targets: TargetCreature("target nonblack creature", NonBlack()),
			Effect:  destroyFirstLegalCardTarget,
		}},
	})
}
