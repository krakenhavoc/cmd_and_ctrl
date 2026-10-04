package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Minamo Scrollkeeper — Creature — Human Wizard {1}{U}, 2/3:
//
//	"Defender
//	 Your maximum hand size is increased by one."
//
// The increase is a maximum-hand-size static (ADR 0113 §3, #2074),
// folded in CR 613.11 timestamp order: a Null Profusion then a
// Scrollkeeper is three, the other order is two (the 2013-04-15
// ruling).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9ad4f876-e923-4a09-8e70-eb826aacf89d",
		Name:            "Minamo Scrollkeeper",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		HandSize:        []game.HandSizeStatic{YourMaxHandSizeChangedBy(1)},
	})
}
