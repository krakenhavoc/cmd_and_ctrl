package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rubblebelt Recluse — Creature — Ogre Berserker {4}{R}, 6/5 (Ravnica
// Allegiance, #5521):
//
//	"This creature attacks each combat if able."
//
// Bloodrock Cyclops's attack requirement (CR 508.1d), the shared
// AttacksEachCombat static.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3ef0fff2-6010-42de-a7ce-e4cac295a481",
		Name:         "Rubblebelt Recluse",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{AttacksEachCombat()},
	})
}
