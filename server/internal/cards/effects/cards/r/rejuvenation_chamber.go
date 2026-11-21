package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rejuvenation Chamber — Artifact {3}:
//
//	"Fading 2 (This artifact enters with two fade counters on it. At the
//	 beginning of your upkeep, remove a fade counter from it. If you
//	 can't, sacrifice it.)
//	 {T}: You gain 2 life."
//
// Fading is Fading/FadingUpkeep (fading.go, CR 702.32a); the life is an
// ordinary tap ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c1ebe189-db71-47cd-a763-8cb8c305383a",
		Name:         "Rejuvenation Chamber",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.Fading(2)},
		Triggered:    []game.TriggeredAbility{effects.FadingUpkeep()},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: You gain 2 life.",
			Purpose: game.Purpose{LifeGain: 2, Answers: game.AnswerValue},
			Cost:    effects.TapCost(),
			Effect:  effects.Do(effects.GainLife{Amount: 2}),
		}},
	})
}
