package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "81bfbc5a-c448-4124-ac25-79bd0944136d",
		Name:         "Sparring Collar",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.GrantToAttached("first strike"),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{R}{R}: Attach this Equipment to target creature you control.",
				Cost:    effects.ManaCost("{R}{R}"),
				Targets: effects.TargetCreature("target creature you control", effects.YouControl()),
				Effect:  effects.AttachSourceToTarget,
			},
			effects.EquipAbility("{1}"),
		},
	})
}
