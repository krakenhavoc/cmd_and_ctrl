package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bramble Armor — Artifact — Equipment {1}{G} (Innistrad: Midnight Hunt):
//
//	"When this Equipment enters, attach it to target creature you control.
//	 Equipped creature gets +2/+1.
//	 Equip {4}"
//
// The entry trigger is Baseball Bat's attach (AttachSourceToTarget): it
// targets on the stack, and a target that has gone gets nothing
// (CR 608.2b). The pump is PumpAttached.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "25328cab-c7c0-4c55-99bb-a4149f6c511a",
		Name:         "Bramble Armor",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(2, 1)},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Bramble Armor — attach it to target creature you control", effects.AttachSourceToTarget),
				effects.PermanentYouControl("target creature you control", effects.Creature()),
			),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{4}")},
	})
}
