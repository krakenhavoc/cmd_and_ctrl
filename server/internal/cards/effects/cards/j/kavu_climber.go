package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kavu Climber — Creature — Kavu {3}{G}{G}, 3/3 (Invasion, S82 #5616):
//
//	"When this creature enters, draw a card."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "91320afd-1d42-4cb5-ae40-4eed2fa91dfe",
		Name:         "Kavu Climber",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Kavu Climber — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
