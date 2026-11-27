package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Deepwood Ghoul — Creature — Zombie {2}{B}, 2/1 (Eighth Edition, S82 OLD batch
// 609, #5607):
//
//	"Pay 2 life: Regenerate this creature."
//
// Sentry of the Underworld's regenerate row (RegenerateThisSource, CR 701.19)
// with only the life payment as its cost; the life is paid as the ability is
// activated, so a player at 1 life cannot.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d99cd2d0-32b8-4edc-8089-557c48dd4be0",
		Name:         "Deepwood Ghoul",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Pay 2 life: Regenerate this creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.PayLife(2),
			Effect:  effects.RegenerateThisSource,
		}},
	})
}
