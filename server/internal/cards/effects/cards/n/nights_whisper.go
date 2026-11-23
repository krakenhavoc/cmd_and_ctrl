package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Night's Whisper — Sorcery {1}{B}:
//
//	"You draw two cards and you lose 2 life."
//
// Sign in Blood without the target — the caster always pays and
// always draws, which is why it costs one less. Same DrawCards +
// ChangePlayerLife composition, aimed at the controller.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7ffae8f8-3006-4969-a339-6d30678f87ea",
		Name:         "Night's Whisper",
		Purpose:      game.Purpose{Draws: 2},
		Completeness: effects.CompletenessFull,
		OnResolve: effects.ResolveSteps(
			// The life loss waits for a prompt the draw asks (CR 608.2c, #4155).
			effects.DrawCards{N: 2},
			effects.GainLife{Amount: -2},
		),
	})
}
