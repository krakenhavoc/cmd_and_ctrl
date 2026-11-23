package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scavenged Blade — Artifact — Equipment {1}{R} (Foundations Jumpstart):
//
//	"When this Equipment enters, attach it to target creature you control.
//	 Equipped creature gets +2/+0.
//	 Equip {2}{R}"
//
// Dwarven Mattock's attach without the Dwarf bar: a targeted enters trigger
// through AttachSourceToTarget. With no creature to name the trigger is
// removed from the stack and the Blade stays unattached (CR 603.3d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2e7b4df3-7c36-469f-8db8-b3081a2c5691",
		Name:         "Scavenged Blade",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(2, 0)},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Scavenged Blade — attach it to target creature you control", effects.AttachSourceToTarget),
				effects.PermanentYouControl("target creature you control", effects.Creature()),
			),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}{R}")},
	})
}
