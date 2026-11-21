package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thought Devourer — Creature — Beast {2}{U}{U}, 4/4:
//
//	"Flying
//	 Your maximum hand size is reduced by four."
//
// The reduction is a maximum-hand-size static (ADR 0113 §3, #2074),
// folded in CR 613.11 timestamp order. Below zero reads as zero
// (CR 107.1b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e5106a2d-caca-4519-a1a1-eabaec4005df",
		Name:            "Thought Devourer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		HandSize:        []game.HandSizeStatic{effects.YourMaxHandSizeChangedBy(-4)},
	})
}
