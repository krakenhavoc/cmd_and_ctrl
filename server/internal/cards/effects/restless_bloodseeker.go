package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Restless Bloodseeker // Bloodsoaked Reveler — a Crimson Vow transformer:
//
//	Restless Bloodseeker {1}{B}, Creature — Vampire, 1/3
//	  "At the beginning of your end step, if you gained life this turn,
//	   create a Blood token.
//	   Sacrifice two Blood tokens: Transform this creature. Activate only
//	   as a sorcery."
//
//	Bloodsoaked Reveler, Creature — Vampire, 3/3
//	  "At the beginning of your end step, if you gained life this turn,
//	   create a Blood token.
//	   {4}{B}: Each opponent loses 2 life and you gain 2 life."
//
// Both faces share the end-step trigger (vowB04BloodAtEndStepIfLifeGained,
// intervening "if" rechecked at resolution, CR 603.4). The transform cost
// is two Blood tokens (refused with fewer) and sorcery timing; it turns
// the permanent over in place (Captive Weird's shape).
//
// No simplification.
func init() {
	const oracle = "7aef5d2e-b66d-4822-b563-44d21d008272"
	Register(Spec{
		OracleID:     oracle,
		Name:         "Restless Bloodseeker",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			vowB04BloodAtEndStepIfLifeGained("Restless Bloodseeker"),
		},
		Activated: []ActivatedAbility{{
			Label:        "Sacrifice two Blood tokens: Transform this creature. Activate only as a sorcery.",
			Cost:         SacrificeN(2, "two Blood tokens", HasSubtype("Blood")),
			SorcerySpeed: true,
			Effect:       Do(TransformThis{}),
		}},
	})
	Register(Spec{
		OracleID:     oracle + "#1",
		Name:         "Bloodsoaked Reveler",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			vowB04BloodAtEndStepIfLifeGained("Bloodsoaked Reveler"),
		},
		Activated: []ActivatedAbility{{
			Label:   "{4}{B}: Each opponent loses 2 life and you gain 2 life.",
			Purpose: game.Purpose{Answers: game.AnswerValue, LifeGain: 2},
			Cost:    ManaCost("{4}{B}"),
			Effect:  Do(EachOpponentLosesLife{N: 2}, GainLife{Amount: 2}),
		}},
	})
}
