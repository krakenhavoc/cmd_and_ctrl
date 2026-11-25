package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jwar Isle Avenger — Creature — Sphinx {4}{U}, 3/3 (Oath of the Gatewatch,
// #5545):
//
//	"Surge {2}{U} (You may cast this spell for its surge cost if you or a
//	 teammate has cast another spell this turn.)
//	 Flying"
//
// Surge is the shared gated alternative cost (CR 702.117a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "2db8fc61-13af-4227-83f9-9f426ab978d3",
		Name:             "Jwar Isle Avenger",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"flying"},
		AlternativeCosts: []game.AlternativeCost{effects.Surge("{2}{U}")},
	})
}
