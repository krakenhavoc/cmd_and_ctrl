package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Horned Helm — Artifact — Equipment {2} (Fifth Dawn, #5604):
//
//	"Equipped creature gets +1/+1 and has trample.
//	 {G}{G}: Attach this Equipment to target creature you control.
//	 Equip {1} ({1}: Attach to target creature you control. Equip only
//	 as a sorcery.)"
//
// The +1/+1 and trample is the ordinary Equipment static. The {G}{G} ability is
// Cranial Plating's attach: no sorcery-speed limit, so it moves the Equipment
// at instant speed, and only onto a creature you control.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "127b438d-e5c2-47ba-9554-d9d5af018d09",
		Name:         "Horned Helm",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 1),
			effects.GrantToAttached("trample"),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{G}{G}: Attach this Equipment to target creature you control.",
				Cost:    effects.ManaCost("{G}{G}"),
				Targets: effects.TargetCreature("target creature you control", effects.YouControl()),
				Effect:  effects.AttachSourceToTarget,
			},
			effects.EquipAbility("{1}"),
		},
	})
}
