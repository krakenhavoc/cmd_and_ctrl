package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crude Rampart — Creature — Wall {3}{W}, 4/5 (Onslaught):
//
//	"Defender
//	 Morph {4}{W}"
//
// Face down it is a 2/2 with no text (CR 708.2a), so it CAN attack
// while it hides; turned face up for {4}{W} it is the 4/5 Wall with
// defender.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "c9823e5f-d0f8-4d28-8b3b-55f94ea9d76f",
		Name:             "Crude Rampart",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"defender"},
		AlternativeCosts: []game.AlternativeCost{effects.Morph("{4}{W}")},
	})
}
