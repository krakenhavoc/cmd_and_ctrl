package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Duskhunter Bat — Creature — Bat {1}{B}, 1/1 (Magic 2012, #5573):
//
//	"Bloodthirst 1 (If an opponent was dealt damage this turn, this
//	 creature enters with a +1/+1 counter on it.)
//	 Flying"
//
// Bloodthirst is the shared entry replacement (bloodthirst.go); flying is
// a printed keyword.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6082f84a-3af2-4007-ab88-7b4b59b5c154",
		Name:            "Duskhunter Bat",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements:    []game.ReplacementEffect{Bloodthirst(1, "Duskhunter Bat")},
	})
}
