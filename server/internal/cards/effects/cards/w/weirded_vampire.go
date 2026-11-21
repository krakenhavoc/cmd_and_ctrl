package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Weirded Vampire — Creature — Vampire Horror {3}{B}, 3/3 (Eldritch Moon,
// #5541):
//
//	"Madness {2}{B} (If you discard this card, discard it into exile. When
//	 you do, cast it for its madness cost or put it into your graveyard.)"
//
// Arrogant Wurm's shape: the one Madness string grows the discard
// replacement and the exile trigger (CR 702.35a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3af07cd3-36c3-4792-bc93-3ec82e1606bb",
		Name:         "Weirded Vampire",
		Completeness: effects.CompletenessFull,
		Madness:      "{2}{B}",
	})
}
