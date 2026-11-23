package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "cee06ea6-8bad-4858-b78e-04879a825bbe",
		Name:         "Neurok Stealthsuit",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.GrantToAttached("shroud"),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{U}{U}: Attach this Equipment to target creature you control.",
				Cost:    effects.ManaCost("{U}{U}"),
				Targets: effects.TargetCreature("target creature you control", effects.YouControl()),
				Effect:  effects.AttachSourceToTarget,
			},
			effects.EquipAbility("{1}"),
		},
	})
}
