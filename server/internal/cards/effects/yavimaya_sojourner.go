package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Yavimaya Sojourner — Creature — Treefolk {7}{G}, 4/6:
//
//	"Domain — This spell costs {1} less to cast for each basic land type
//	 among lands you control."
//
// Leyline Binding's self cost modifier on the shared domain count
// (DomainCostReduction, #5666). Generic mana only (CR 601.2f): with five
// types it costs {G}.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b3e9917e-1bcf-4e8e-9550-7f5897dcb433",
		Name:         "Yavimaya Sojourner",
		Completeness: CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			DomainCostReduction("Domain — This spell costs {1} less to cast for each basic land type among lands you control"),
		},
	})
}
