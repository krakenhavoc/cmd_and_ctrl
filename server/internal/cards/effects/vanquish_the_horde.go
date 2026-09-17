package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vanquish the Horde — Sorcery {6}{W}{W}:
//
//	"This spell costs {1} less to cast for each creature on the
//	 battlefield.
//	 Destroy all creatures."
//
// #746: the Blasphemous Act reduction, on a Wrath.
func init() {
	Register(Spec{
		OracleID:     "a332e80a-dc51-4dc6-bc85-e114a1c6fdb8",
		Name:         "Vanquish the Horde",
		Completeness: CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			CostsLessEach(PermanentsOnBattlefield(Creature()),
				"This spell costs {1} less to cast for each creature on the battlefield."),
		},
		OnResolve: wrathDestroyAllCreatures,
	})
}
