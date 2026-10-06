package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gorger Wurm — "Devour 1"
//
// Devour 1 is devour.go's: sacrifice any number of creatures as it
// enters, one +1/+1 counters for each. No simplifications.
func init() {
	Register(Spec{
		OracleID:     "f2126eb9-892c-4788-95cc-6f3e59f0375d",
		Name:         "Gorger Wurm",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{Devour("Gorger Wurm", 1)},
	})
}
