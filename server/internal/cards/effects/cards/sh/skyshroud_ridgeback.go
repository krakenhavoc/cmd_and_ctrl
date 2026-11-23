package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skyshroud Ridgeback — Creature — Beast {G}, 2/3:
//
//	"Fading 2 (This creature enters with two fade counters on it. At the
//	 beginning of your upkeep, remove a fade counter from it. If you
//	 can't, sacrifice it.)"
//
// Fading is Fading/FadingUpkeep (fading.go, CR 702.32a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "34a160bc-5d7e-483c-b24c-ab30b7d313b0",
		Name:         "Skyshroud Ridgeback",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.Fading(2)},
		Triggered:    []game.TriggeredAbility{effects.FadingUpkeep()},
	})
}
