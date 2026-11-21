package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mechanical Glider — Artifact — Equipment {1} (Avatar: The Last
// Airbender Eternal, #5075):
//
//	"When this Equipment enters, attach it to target creature you control.
//	 Equipped creature has flying. (It can't be blocked except by
//	 creatures with flying or reach.)
//	 Equip {2} ({2}: Attach to target creature you control. Equip only as
//	 a sorcery.)"
//
// Falcon's Wing Harness's shape: the entry trigger targets a creature you
// control and attaches the Equipment to it if it is still legal
// (CR 608.2b); the flying is a layer-6 grant to the equipped creature;
// equip is CR 702.6a's sorcery-speed activated ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d53d1805-c4ec-4b42-a226-3cd78cd2be84",
		Name:         "Mechanical Glider",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.GrantToAttached("flying"),
		},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Mechanical Glider — attach it to target creature you control", effects.AttachSourceToTarget),
				effects.PermanentYouControl("target creature you control", effects.Creature()),
			),
		},
		Activated: []effects.ActivatedAbility{
			effects.EquipAbility("{2}"),
		},
	})
}
