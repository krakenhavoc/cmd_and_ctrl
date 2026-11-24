package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Time Sieve — Artifact {U}{B} (Secret Lair Drop):
//
//	"{T}, Sacrifice five artifacts: Take an extra turn after this one."
//
// Five artifacts are the sacrifice cost, chosen as the ability is activated
// (Time Sieve may be one of them, it is an artifact too); the extra turn is
// TakeExtraTurn (CR 500.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3da5977a-36d4-4f32-ab9b-8b93809d818d",
		Name:         "Time Sieve",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Sacrifice five artifacts: Take an extra turn after this one.",
			Purpose: game.Purpose{ExtraTurns: 1, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.TapCost(), effects.SacrificeN(5, "five artifacts", effects.Artifact())),
			Effect:  effects.Do(effects.TakeExtraTurn{}),
		}},
	})
}
