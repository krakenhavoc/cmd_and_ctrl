package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thought Eater — Creature — Beast {1}{U}, 2/2:
//
//	"Flying
//	 Your maximum hand size is reduced by three."
//
// The reduction is a maximum-hand-size static (ADR 0113 §3, #2074),
// folded in CR 613.11 timestamp order. Below zero reads as zero
// (CR 107.1b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "88e18dc9-8b11-4369-9860-a687925cdcb4",
		Name:            "Thought Eater",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		HandSize:        []game.HandSizeStatic{YourMaxHandSizeChangedBy(-3)},
	})
}
