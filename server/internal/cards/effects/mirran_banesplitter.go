package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mirran Banesplitter — Artifact — Equipment {R} (March of the Machine):
//
//	"Flash
//	 When this Equipment enters, attach it to target creature you control.
//	 Equipped creature gets +2/+0.
//	 Equip {3}"
//
// Baseball Bat's attach trigger (AttachSourceToTarget) and PumpAttached.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "aeaf3ce8-6bdf-4a90-8b81-cdbce0f3bb12",
		Name:            "Mirran Banesplitter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Static:          []game.StaticAbility{PumpAttached(2, 0)},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Mirran Banesplitter — attach it to target creature you control", AttachSourceToTarget),
				PermanentYouControl("target creature you control", Creature()),
			),
		},
		Activated: []ActivatedAbility{EquipAbility("{3}")},
	})
}
