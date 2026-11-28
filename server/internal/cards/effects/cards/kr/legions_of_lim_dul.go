package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Legions of Lim-Dûl — Creature — Zombie {1}{B}{B}, 2/3 (Ice Age, S82 OLD, #5636):
//
//	"Snow swampwalk (This creature can't be blocked as long as defending player controls a snow Swamp.)"
//
// The snow landwalk keyword (CR 702.14), as Rime Dryad's with Swamps.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "70b69803-1870-4441-954b-dba5b433fcb2",
		Name:            "Legions of Lim-Dûl",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"snow swampwalk"},
	})
}
