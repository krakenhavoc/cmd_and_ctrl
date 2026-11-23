package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thunder-Thrash Elder — "Devour 3"
//
// Devour 3 is devour.go's: sacrifice any number of creatures as it
// enters, three +1/+1 counters for each. No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "aef6c5b3-824f-4b7d-9671-1ed61d547717",
		Name:         "Thunder-Thrash Elder",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.Devour("Thunder-Thrash Elder", 3)},
	})
}
