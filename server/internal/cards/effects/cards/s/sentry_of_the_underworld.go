package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sentry of the Underworld — Creature — Griffin Skeleton {3}{W}{B}, 3/3 (Theros, S82 OLD b391, #5561):
//
//	"Flying, vigilance
//	 {W}{B}, Pay 3 life: Regenerate this creature."
//
// Mire Boa's regenerate row (RegenerateThisSource, CR 701.19) with the
// life payment as a second cost; the life is paid as the ability is
// activated, so a player at 2 life cannot.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fd4d83fa-070c-46bf-a620-041f69acdea9",
		Name:            "Sentry of the Underworld",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}{B}, Pay 3 life: Regenerate this creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.Plus(effects.ManaCost("{W}{B}"), effects.PayLife(3)),
			Effect:  effects.RegenerateThisSource,
		}},
	})
}
