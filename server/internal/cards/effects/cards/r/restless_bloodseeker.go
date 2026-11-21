package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Restless Bloodseeker",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.VowB04BloodAtEndStepIfLifeGained("Restless Bloodseeker"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:        "Sacrifice two Blood tokens: Transform this creature. Activate only as a sorcery.",
			Cost:         effects.SacrificeN(2, "two Blood tokens", effects.HasSubtype("Blood")),
			SorcerySpeed: true,
			Effect:       effects.Do(effects.TransformThis{}),
		}},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Bloodsoaked Reveler",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.VowB04BloodAtEndStepIfLifeGained("Bloodsoaked Reveler"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}{B}: Each opponent loses 2 life and you gain 2 life.",
			Purpose: game.Purpose{Answers: game.AnswerValue, LifeGain: 2},
			Cost:    effects.ManaCost("{4}{B}"),
			Effect:  effects.Do(effects.EachOpponentLosesLife{N: 2}, effects.GainLife{Amount: 2}),
		}},
	})
}
