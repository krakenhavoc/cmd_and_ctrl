package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thought Nibbler — Creature — Beast {U}, 1/1:
//
//	"Flying
//	 Your maximum hand size is reduced by two."
//
// The reduction is a maximum-hand-size static (ADR 0113 §3, #2074),
// folded in CR 613.11 timestamp order. Below zero reads as zero
// (CR 107.1b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9a640695-fc17-4ccc-ac88-ce48b6c6ad53",
		Name:            "Thought Nibbler",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		HandSize:        []game.HandSizeStatic{effects.YourMaxHandSizeChangedBy(-2)},
	})
}
