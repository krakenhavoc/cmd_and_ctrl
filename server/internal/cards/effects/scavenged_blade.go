package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "2e7b4df3-7c36-469f-8db8-b3081a2c5691",
		Name:         "Scavenged Blade",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(2, 0)},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Scavenged Blade — attach it to target creature you control", AttachSourceToTarget),
				PermanentYouControl("target creature you control", Creature()),
			),
		},
		Activated: []ActivatedAbility{EquipAbility("{2}{R}")},
	})
}
