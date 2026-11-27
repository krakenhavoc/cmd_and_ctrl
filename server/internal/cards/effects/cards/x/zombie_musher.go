package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Zombie Musher — Snow Creature — Zombie {3}{B}, 2/3 (Coldsnap):
//
//	"Snow landwalk
//	 {S}: Regenerate this creature."
//
// Snow landwalk is an engine keyword (CR 702.14c): it can't be blocked while
// the defending player controls a snow land. The {S} is paid with snow mana
// only (CR 107.4h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "89f87341-c18f-4dc7-ba2e-81305383231e",
		Name:            "Zombie Musher",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"snow landwalk"},
		Activated:       []effects.ActivatedAbility{effects.LrwB1RegenerateThisRow("{S}")},
	})
}
