package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "6d2cf8b4-1468-4e00-94ca-65ef0c564b09",
		Name:         "Skyclave Pick-Axe",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Skyclave Pick-Axe — attach it to target creature you control", AttachSourceToTarget),
				PermanentYouControl("target creature you control", Creature()),
			),
			Landfall("Skyclave Pick-Axe — equipped creature gets +2/+2 until end of turn",
				Do(Instruction(func(ctx *Context) error {
					return pumpEquippedCreatureUntilEOT(ctx, 2, 2, "Skyclave Pick-Axe — +2/+2")
				}))),
		},
		Activated: []ActivatedAbility{EquipAbility("{2}{G}")},
	})
}
