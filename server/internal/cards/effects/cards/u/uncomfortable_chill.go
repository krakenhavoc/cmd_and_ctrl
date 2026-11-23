package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Uncomfortable Chill — Instant {2}{U} (Foundations Jumpstart, #5929):
//
//	"Creatures your opponents control get -2/-0 until end of turn.
//	 Draw a card."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d6596336-8177-4475-95b9-a21858563276",
		Name:         "Uncomfortable Chill",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		OnResolve: effects.ResolveSteps(
			effects.BoostUntilEOT{
				Match: effects.And(effects.Creature(), effects.OpponentControls()), Power: -2,
				Label: "Uncomfortable Chill — creatures your opponents control get -2/-0",
			},
			effects.DrawCards{N: 1},
		),
	})
}
