package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Devotee of Strength — Creature — Snake Wizard {2}{G}, 3/2 (Hour of
// Devastation, #5534):
//
//	"{4}{G}: Target creature gets +2/+2 until end of turn."
//
// Deepwood Drummer's row without its tap and discard: the shared
// boostTargetUntilEOT body at +2/+2, declared a benefit to the target. A
// target that left in response is skipped (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a6d8f944-f516-49d7-9aa8-77b98e862d6a",
		Name:         "Devotee of Strength",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label: "{4}{G}: Target creature gets +2/+2 until end of turn.",
			Purpose: game.Purpose{
				Answers: game.AnswerPump,
				Targets: game.ForTargets(effects.BenefitUntilEOT(0, 2, 2)),
			},
			Cost:    effects.ManaCost("{4}{G}"),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.BoostTargetUntilEOT("Devotee of Strength — +2/+2", 2, 2),
		}},
	})
}
