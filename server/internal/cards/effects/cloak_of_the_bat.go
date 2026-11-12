package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cloak of the Bat — Artifact — Equipment {2} (Commander Legends: Battle
// for Baldur's Gate, #5490):
//
//	"Equipped creature has flying and haste.
//	 Equip {2} ({2}: Attach to target creature you control. Equip only as
//	 a sorcery.)"
//
// Dragonwing Glider's grant without the pump: a layer-6 grant of both
// keywords to the equipped creature, and one equip ability. A creature
// equipped the turn it arrives can attack that turn (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ecd28e2a-e172-4f4c-8416-b9a0d97ee497",
		Name:         "Cloak of the Bat",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{GrantToAttached("flying", "haste")},
		Activated:    []ActivatedAbility{EquipAbility("{2}")},
	})
}
