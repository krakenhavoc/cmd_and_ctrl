package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Messenger Falcons — Creature — Bird {2}{G/U}{W}, 2/2 (Alara Reborn,
// #5586):
//
//	"Flying
//	 When this creature enters, draw a card."
//
// The enters trigger uses the stack (Arctic Wolves' draw).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "01fda82a-227d-4e3d-8481-a13c3812392a",
		Name:            "Messenger Falcons",
		Completeness:    effects.CompletenessFull,
		Purpose:         game.Purpose{Draws: 1},
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Messenger Falcons — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
