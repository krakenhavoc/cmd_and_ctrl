package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stream Hopper — Creature — Goblin {U/R}, 1/1 (Eventide, #5590):
//
//	"{U/R}: This creature gains flying until end of turn."
//
// Killer Whale's self grant at a hybrid cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "153cd93c-7bba-4482-893c-55a602c4439b",
		Name:         "Stream Hopper",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{U/R}: This creature gains flying until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{U/R}"),
			Effect:  effects.ThisGainsKeywordUntilEndOfTurn("flying", "Stream Hopper — flying until end of turn"),
		}},
	})
}
