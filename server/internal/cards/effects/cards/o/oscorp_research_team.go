package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Oscorp Research Team — Creature — Human Scientist {3}{U}, 1/5
// (Marvel's Spider-Man):
//
//	"{6}{U}: Draw two cards."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4d2e233c-0173-417f-82a0-1e692a400ae1",
		Name:         "Oscorp Research Team",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{6}{U}: Draw two cards.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{6}{U}"),
			Effect:  effects.Do(effects.DrawCards{N: 2}),
		}},
	})
}
