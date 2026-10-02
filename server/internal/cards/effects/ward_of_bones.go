package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ward of Bones — Artifact {6}:
//
//	"Each opponent who controls more creatures than you can't cast
//	 creature spells. The same is true for artifacts and enchantments.
//	 Each opponent who controls more lands than you can't play lands."
//
// ADR 0109 §4 (#1895). Four statics, each comparing the opponent's count
// with the Ward's controller's, live at every query (nothing is stored):
// three CastRestrictions ("the same is true for artifacts and
// enchantments" is the same clause with the type changed, so a player
// who controls more artifacts can't cast artifact spells) and one
// LandPlayRestriction. A tie, or fewer than you, restricts nobody.
//
// The counts are permanents each player CONTROLS, not owns, and read the
// layer-resolved types, so an animated land counts as both a creature
// and a land.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c3ea1497-63ac-46bb-95e9-d98b93a880b3",
		Name:         "Ward of Bones",
		Completeness: CompletenessFull,
		CastRestrictions: []game.CastRestriction{
			OpponentsWithMoreCantCast("Each opponent who controls more creatures than you can't cast creature spells.",
				QueryType("creature"), Creature()),
			OpponentsWithMoreCantCast("Each opponent who controls more artifacts than you can't cast artifact spells.",
				QueryType("artifact"), Artifact()),
			OpponentsWithMoreCantCast("Each opponent who controls more enchantments than you can't cast enchantment spells.",
				QueryType("enchantment"), Enchantment()),
		},
		LandPlayRestrictions: []game.LandPlayRestriction{
			OpponentsWithMoreLandsCantPlayLands("Each opponent who controls more lands than you can't play lands."),
		},
	})
}
