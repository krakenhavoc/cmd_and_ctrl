package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Screeching Harpy — Creature — Harpy Beast {2}{B}{B}, 2/2 (Tempest, S82 OLD
// b753, #5628):
//
//	"Flying
//	 {1}{B}: Regenerate this creature."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "58bbd316-e22c-4c23-8024-364122d1e9a2",
		Name:            "Screeching Harpy",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}: Regenerate this creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.ManaCost("{1}{B}"),
			Effect:  effects.RegenerateThisSource,
		}},
	})
}
