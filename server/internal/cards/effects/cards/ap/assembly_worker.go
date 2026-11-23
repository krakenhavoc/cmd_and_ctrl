package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Assembly-Worker — Artifact Creature — Assembly-Worker {3}, 2/2
// (Dominaria Remastered):
//
//	"{T}: Target Assembly-Worker creature gets +1/+1 until end of turn."
//
// Mishra's Factory's second ability, word for word.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "751321ac-af07-4b7e-976f-5bf4e4281470",
		Name:         "Assembly-Worker",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Target Assembly-Worker creature gets +1/+1 until end of turn.",
			Purpose: effects.WithAnswers(effects.ForTargets(effects.BenefitUntilEOT(0, 1, 1)), game.AnswerPump),
			Cost:    game.AbilityCost{Tap: true},
			Targets: effects.TargetCreature("target Assembly-Worker creature", effects.Subtype("Assembly-Worker")),
			Effect:  effects.BoostTargetUntilEOT("Assembly-Worker — +1/+1", 1, 1),
		}},
	})
}
