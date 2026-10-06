package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thunder-Thrash Elder — "Devour 3"
//
// Devour 3 is devour.go's: sacrifice any number of creatures as it
// enters, three +1/+1 counters for each. No simplifications.
func init() {
	Register(Spec{
		OracleID:     "aef6c5b3-824f-4b7d-9671-1ed61d547717",
		Name:         "Thunder-Thrash Elder",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{Devour("Thunder-Thrash Elder", 3)},
	})
}
