package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gigastorm Titan — Creature — Elemental {4}{U}, 4/4 (Edge of Eternities,
// #4461):
//
//	"This spell costs {3} less to cast if you've cast another spell this
//	 turn."
//
// Focus the Mind's reduction on a creature spell: the caster's cast
// tally is read as the cost is determined (CR 601.2f), and the Titan is
// not counted yet, because a spell becomes cast only at CR 601.2i, so
// any spell in the tally is another one.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ef06e10c-553b-45b8-abf0-1d2e87c22128",
		Name:         "Gigastorm Titan",
		Completeness: CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			CostsLess(3, "This spell costs {3} less to cast if you've cast another spell this turn.",
				tdmCastAnotherSpellThisTurn),
		},
	})
}
