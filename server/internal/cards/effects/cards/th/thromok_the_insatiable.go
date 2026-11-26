package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thromok the Insatiable — Legendary Creature — Hellion {3}{R}{G}, 0/0 (Planechase
// 2012, S82 OLD, #5569):
//
//	"Devour X, where X is the number of creatures devoured this way (As this
//	 creature enters, you may sacrifice any number of creatures. It enters with
//	 X +1/+1 counters on it for each of those creatures.)"
//
// The #6267 seam (DevourX): sacrificing k creatures as it enters puts k times k
// +1/+1 counters on it (CR 702.82a with N = k), on the entry event so a counter
// doubler applies; devouring none leaves a 0/0 that dies to state-based actions
// (CR 704.5f).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fbb6ec41-3017-4d8d-bca0-c93040f1b291",
		Name:         "Thromok the Insatiable",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.DevourX("Thromok the Insatiable")},
	})
}
