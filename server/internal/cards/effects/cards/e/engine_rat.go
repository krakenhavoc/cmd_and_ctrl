package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Engine Rat — Creature — Zombie Rat {B}, 1/1 (Aetherdrift, #4464):
//
//	"Deathtouch
//	 {5}{B}: Each opponent loses 2 life."
//
// A repeatable mana sink with no tap: each activation is its own life
// loss for every opponent, not damage, so nothing prevents it.
// Deathtouch is the engine's (CR 702.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7e29aae9-815a-4999-ace6-7bc860acde32",
		Name:            "Engine Rat",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{5}{B}: Each opponent loses 2 life.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{5}{B}"),
			Effect:  effects.Do(effects.EachOpponentLosesLife{N: 2}),
		}},
	})
}
