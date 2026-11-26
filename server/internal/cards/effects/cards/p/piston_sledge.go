package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Piston Sledge — Artifact — Equipment {3} (Mirrodin Besieged, S82 OLD b466,
// #5576):
//
//	"When this Equipment enters, attach it to target creature you control.
//	 Equipped creature gets +3/+1.
//	 Equip—Sacrifice an artifact."
//
// The entry trigger targets on the stack (Barbed Bloodletter's attach); a
// target that has gone gets nothing (CR 608.2b). The equip's cost is a
// sacrifice of any artifact the controller picks as the ability is announced
// (EquipPayingAbility, Demonmail Hauberk's shape), the Sledge itself
// included; the equip is still a sorcery-speed ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ba5d8c0d-3e4e-4652-86d2-80ce2f48dc0f",
		Name:         "Piston Sledge",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(3, 1)},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Piston Sledge — attach it to target creature you control", effects.AttachSourceToTarget),
				effects.PermanentYouControl("target creature you control", effects.Creature()),
			),
		},
		Activated: []effects.ActivatedAbility{
			effects.EquipPayingAbility("Equip—Sacrifice an artifact.", effects.SacrificeN(1, "an artifact", effects.Artifact())),
		},
	})
}
