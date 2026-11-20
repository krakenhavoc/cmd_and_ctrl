package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sparring Collar — Artifact — Equipment {2} (Fifth Dawn, #5604):
//
//	"Equipped creature has first strike.
//	 {R}{R}: Attach this Equipment to target creature you control.
//	 Equip {1} ({1}: Attach to target creature you control. Equip only
//	 as a sorcery.)"
//
// The first strike is the ordinary Equipment static. The {R}{R} ability is
// Cranial Plating's attach: no sorcery-speed limit, so it moves the Equipment
// at instant speed, and only onto a creature you control.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "81bfbc5a-c448-4124-ac25-79bd0944136d",
		Name:         "Sparring Collar",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			GrantToAttached("first strike"),
		},
		Activated: []ActivatedAbility{
			{
				Label:   "{R}{R}: Attach this Equipment to target creature you control.",
				Cost:    ManaCost("{R}{R}"),
				Targets: TargetCreature("target creature you control", YouControl()),
				Effect:  AttachSourceToTarget,
			},
			EquipAbility("{1}"),
		},
	})
}
