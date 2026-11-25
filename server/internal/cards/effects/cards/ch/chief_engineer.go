package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chief Engineer — Creature — Vedalken Artificer {1}{U}, 1/3 (Commander 2016,
// #5537):
//
//	"Artifact spells you cast have convoke. (Your creatures can help cast
//	 those spells. Each creature you tap while casting an artifact spell pays
//	 for {1} or one mana of that creature's color.)"
//
// Dazzling Theater's grant (SpellsYouCastHaveConvoke) over artifact spells,
// read against the card being cast (CR 702.51a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "02dbb6f8-ef84-4156-87dc-34033e64fdb7",
		Name:         "Chief Engineer",
		Completeness: effects.CompletenessFull,
		GrantedTapCosts: []game.GrantedTapCost{effects.SpellsYouCastHaveConvoke(
			"Artifact spells you cast have convoke.", effects.SpellIs(effects.Artifact()))},
	})
}
