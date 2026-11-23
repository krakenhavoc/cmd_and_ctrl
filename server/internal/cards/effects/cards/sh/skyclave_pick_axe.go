package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skyclave Pick-Axe — Artifact — Equipment {G} (Foundations Jumpstart):
//
//	"When this Equipment enters, attach it to target creature you control.
//	 Landfall — Whenever a land you control enters, equipped creature gets
//	 +2/+2 until end of turn.
//	 Equip {2}{G}"
//
// Adventuring Gear's landfall pump (the equipped creature is read as the
// trigger resolves, CR 608.2h) with Scavenged Blade's attach on entering.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6d2cf8b4-1468-4e00-94ca-65ef0c564b09",
		Name:         "Skyclave Pick-Axe",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Skyclave Pick-Axe — attach it to target creature you control", effects.AttachSourceToTarget),
				effects.PermanentYouControl("target creature you control", effects.Creature()),
			),
			effects.Landfall("Skyclave Pick-Axe — equipped creature gets +2/+2 until end of turn",
				effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					return effects.PumpEquippedCreatureUntilEOT(ctx, 2, 2, "Skyclave Pick-Axe — +2/+2")
				}))),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}{G}")},
	})
}
