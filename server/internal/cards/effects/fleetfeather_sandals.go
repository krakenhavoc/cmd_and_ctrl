package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fleetfeather Sandals — Artifact — Equipment {2} (Theros, #5561):
//
//	"Equipped creature has flying and haste.
//	 Equip {2} ({2}: Attach to target creature you control. Equip only as a sorcery.)"
//
// Cloak of the Bat's grant and equip row.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9811d1fa-1398-49a2-bcbe-0e40d9779ef1",
		Name:         "Fleetfeather Sandals",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{GrantToAttached("flying", "haste")},
		Activated:    []ActivatedAbility{EquipAbility("{2}")},
	})
}
