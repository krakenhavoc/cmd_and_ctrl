package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// High Ground — Enchantment {W}:
//
//	"Each creature you control can block an additional creature each
//	 combat."
//
// CanBlockAdditional over "creatures you control" (#1706). The effect
// is High Ground's, so a creature of yours that has lost all its
// abilities still blocks two; two High Grounds let each block three.
func init() {
	Register(Spec{
		OracleID:     "b59053d5-b5f2-44aa-a29c-7f3908d86577",
		Name:         "High Ground",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{CanBlockAdditional(b16CreaturesYouControl, 1)},
	})
}
