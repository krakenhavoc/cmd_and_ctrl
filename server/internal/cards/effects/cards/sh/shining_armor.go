package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shining Armor — Artifact — Equipment {1}{W} (Dominaria Remastered):
//
//	"Flash
//	 When this Equipment enters, attach it to target Knight you control.
//	 Equipped creature gets +0/+2 and has vigilance.
//	 Equip {3}"
//
// Bramble Armor's attach with a Knight target, and Barbed Bloodletter's
// flash.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "090a5f76-206e-4b3b-bc1d-9c9706d9955d",
		Name:            "Shining Armor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Static:          []game.StaticAbility{effects.PumpAttached(0, 2), effects.GrantToAttached("vigilance")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Shining Armor — attach it to target Knight you control", effects.AttachSourceToTarget),
				effects.PermanentYouControl("target Knight you control", effects.OfCreatureType("Knight")),
			),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{3}")},
	})
}
