package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tethered Griffin — Creature {W}, 2/3:
//
//	"Flying
//	 When you control no enchantments, sacrifice this creature."
//
// ADR 0107 §1 (#1858). Flying is the printed keyword. The sacrifice is a
// CR 603.8 state trigger over the enchantments its controller controls.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "909e1bff-237a-4259-9c0f-419185681782",
		Name:         "Tethered Griffin",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenYouControlNo(QueryType("enchantment"), "Tethered Griffin — sacrifice it", SacrificeThisIfStillOnBattlefield),
		},
	})
}
