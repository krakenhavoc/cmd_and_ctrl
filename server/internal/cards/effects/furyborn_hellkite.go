package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Furyborn Hellkite — Creature — Dragon {4}{R}{R}{R}, 6/6 (Magic 2012,
// #5573):
//
//	"Bloodthirst 6 (If an opponent was dealt damage this turn, this
//	 creature enters with six +1/+1 counters on it.)
//	 Flying"
//
// Bloodthirst is the shared entry replacement (bloodthirst.go); flying is
// a printed keyword.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8c4aba38-7fbd-463e-8454-21037090c71e",
		Name:            "Furyborn Hellkite",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements:    []game.ReplacementEffect{Bloodthirst(6, "Furyborn Hellkite")},
	})
}
