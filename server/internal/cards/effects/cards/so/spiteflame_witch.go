package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Spiteflame Witch — Creature — Elemental Shaman {1}{B/R}, 2/1 (Shadowmoor,
// S82 OLD b529, #5591):
//
//	"{B}{R}: Each player loses 1 life."
//
// An activated ability with no target: every seated player, the Witch's
// controller included, loses 1 life as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b4ab38e5-c446-4845-b509-4d6f4438949d",
		Name:         "Spiteflame Witch",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}{R}: Each player loses 1 life.",
			Cost:    effects.ManaCost("{B}{R}"),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect:  effects.NemEachPlayerLoses(1),
		}},
	})
}
