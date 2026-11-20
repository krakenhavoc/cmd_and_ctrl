package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Neurok Stealthsuit — Artifact — Equipment {2} (Fifth Dawn, #5604):
//
//	"Equipped creature has shroud. (It can't be the target of spells or
//	 abilities.)
//	 {U}{U}: Attach this Equipment to target creature you control.
//	 Equip {1} ({1}: Attach to target creature you control. Equip only
//	 as a sorcery.)"
//
// The shroud is the ordinary Equipment static. The {U}{U} ability is
// Cranial Plating's attach: no sorcery-speed limit, so it moves the Equipment
// at instant speed, and only onto a creature you control.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cee06ea6-8bad-4858-b78e-04879a825bbe",
		Name:         "Neurok Stealthsuit",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			GrantToAttached("shroud"),
		},
		Activated: []ActivatedAbility{
			{
				Label:   "{U}{U}: Attach this Equipment to target creature you control.",
				Cost:    ManaCost("{U}{U}"),
				Targets: TargetCreature("target creature you control", YouControl()),
				Effect:  AttachSourceToTarget,
			},
			EquipAbility("{1}"),
		},
	})
}
