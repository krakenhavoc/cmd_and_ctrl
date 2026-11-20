package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Carnage Wurm — Creature — Wurm {6}{G}, 6/6 (Magic 2012, #5573):
//
//	"Bloodthirst 3 (If an opponent was dealt damage this turn, this
//	 creature enters with three +1/+1 counters on it.)
//	 Trample (This creature can deal excess combat damage to the player
//	 or planeswalker it's attacking.)"
//
// Bloodthirst is the shared entry replacement (bloodthirst.go); trample is
// a printed keyword.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "007161cb-a44b-4443-a53a-dfaacc6cfb86",
		Name:            "Carnage Wurm",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Replacements:    []game.ReplacementEffect{Bloodthirst(3, "Carnage Wurm")},
	})
}
