package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Haunted Cloak — Artifact — Equipment for {3} (EDHREC rank 2127):
//
//	"Equipped creature has vigilance, trample, and haste.
//	 Equip {1}"
//
// Three keywords, no stat line, and an equip of one — the cheapest
// way in the catalog to move haste and vigilance onto a new body
// every turn. Basilisk Collar's shape in a different colour of
// keyword.
//
// The equip cost is the card: at {1} it re-equips the turn the
// carrier dies, which is what separates an Equipment from an Aura and
// what a Voltron deck is actually buying.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "74fc88d5-19c3-4516-9a06-e9573d659caa",
		Name:         "Haunted Cloak",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			GrantToAttached("vigilance", "trample", "haste"),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{1}"),
		},
	})
}
