package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tunneler Wurm — Creature — Wurm {6}{G}{G}, 6/6 (Judgment, #5611):
//
//	"Discard a card: Regenerate this creature."
//
// Rootwater Alligator's regeneration with a discard for its cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "31370fdf-c666-40d2-a7ff-3ba1dfe350cd",
		Name:         "Tunneler Wurm",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Discard a card: Regenerate this creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.DiscardACard(),
			Effect:  effects.RegenerateThisSource,
		}},
	})
}
