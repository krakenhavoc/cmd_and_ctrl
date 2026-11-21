package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Succumb to Temptation — Instant {1}{B}{B} (Eldritch Moon, #5541):
//
//	"You draw two cards and you lose 2 life."
//
// Night's Whisper's text at instant speed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c8e21bc4-ea57-46d6-9e7e-ad95e17ea0fd",
		Name:         "Succumb to Temptation",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 2},
		OnResolve: effects.ResolveSteps(
			// The life loss waits for a prompt the draw asks (CR 608.2c, #4155).
			effects.DrawCards{N: 2},
			effects.GainLife{Amount: -2},
		),
	})
}
