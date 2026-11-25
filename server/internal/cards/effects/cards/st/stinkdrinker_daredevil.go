package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stinkdrinker Daredevil — Creature — Goblin Rogue {2}{R}, 1/3 (Commander 2015,
// S82 OLD #5546):
//
//	"Giant spells you cast cost {2} less to cast."
//
// Knight of the Stampede's reduction for Giants: a CostReduction over your own
// Giant spells, applied at CR 601.2f against generic mana only, so the coloured
// pips are never touched. An opponent's Giant spell is not discounted.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "25af5e13-ebe8-4bc4-b8fe-8b9aee59579c",
		Name:         "Stinkdrinker Daredevil",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.CostsLess(2, "Giant spells you cast cost {2} less to cast.",
				effects.YourSpell(), effects.SpellOfCreatureType("Giant")),
		},
	})
}
