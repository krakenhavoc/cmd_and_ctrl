package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "aeaf3ce8-6bdf-4a90-8b81-cdbce0f3bb12",
		Name:            "Mirran Banesplitter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Static:          []game.StaticAbility{effects.PumpAttached(2, 0)},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Mirran Banesplitter — attach it to target creature you control", effects.AttachSourceToTarget),
				effects.PermanentYouControl("target creature you control", effects.Creature()),
			),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{3}")},
	})
}
