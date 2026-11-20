package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cyclops of Eternal Fury — Enchantment Creature — Cyclops {4}{R}{R}, 5/3 (Journey into Nyx, #5558):
//
//	"Creatures you control have haste."
//
// Fervor's haste grant on a creature: a layer-6 grant to every creature
// its controller controls, itself included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f7875f87-d774-4811-a4a7-349ec01bf95b",
		Name:         "Cyclops of Eternal Fury",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalKeywordGrant(TribeFilter{YoursOnly: true}, "haste"),
		},
	})
}
