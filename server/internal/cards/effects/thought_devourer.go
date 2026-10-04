package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "e5106a2d-caca-4519-a1a1-eabaec4005df",
		Name:            "Thought Devourer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		HandSize:        []game.HandSizeStatic{YourMaxHandSizeChangedBy(-4)},
	})
}
