package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Henzie "Toolbox" Torre — Legendary Creature — Devil Rogue {B}{R}{G},
// 3/3 (New Capenna Commander, #5667):
//
//	"Each creature spell you cast with mana value 4 or greater has blitz.
//	 The blitz cost is equal to its mana cost. (You may choose to cast
//	 that spell for its blitz cost. If you do, it gains haste and "When
//	 this creature dies, draw a card." Sacrifice it at the beginning of
//	 the next end step.)
//	 Blitz costs you pay cost {1} less for each time you've cast your
//	 commander from the command zone this game."
//
// The first ability is BlitzForCreatureSpellsYouCast (granted_blitz.go):
// each creature spell its controller casts whose mana value is 4 or more
// may be cast for a blitz cost equal to its own mana cost, from any zone
// where that mana cost could be paid (CR 118.9a). The permanent it
// becomes then has haste and "When this creature dies, draw a card." for
// as long as it stays on the battlefield and is sacrificed at the
// beginning of the next end step (CR 702.152a). Those hang on the blitz
// cost having been paid, so they happen even if Henzie has left by the
// time the spell resolves. A creature that prints blitz keeps its own
// cost beside this one (CR 702.152b).
//
// The second is a reduction of blitz costs alone
// (BlitzCostsYouPayCostLessForEachCommanderCast): the granted cost and a
// card's printed one, never the same spell cast for its mana cost. It
// takes generic mana only (CR 118.7a) and is applied with the other
// reductions, after a commander's tax (CR 601.2f), so Henzie's own
// commander paying a blitz cost from the command zone has its tax eaten
// by the reduction. Every cast of a commander its controller owns from
// the command zone counts, whichever commander it was.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5320f5a6-8066-47d4-aa0a-4c879f39d413",
		Name:         "Henzie \"Toolbox\" Torre",
		Completeness: CompletenessFull,
		GrantedAlternativeCosts: []game.GrantedAlternativeCost{
			BlitzForCreatureSpellsYouCast(4),
		},
		CostModifiers: []game.CostModifier{
			BlitzCostsYouPayCostLessForEachCommanderCast(),
		},
	})
}
