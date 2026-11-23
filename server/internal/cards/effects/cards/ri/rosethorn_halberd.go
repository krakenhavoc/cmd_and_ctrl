package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rosethorn Halberd — Artifact — Equipment {G} (Lord of the Rings
// Commander):
//
//	"When this Equipment enters, attach it to target non-Human creature
//	 you control.
//	 Equipped creature gets +2/+1.
//	 Equip {5}"
//
// Bramble Armor's attach with a non-Human target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "94848b02-757c-456c-a12c-b0be193221d0",
		Name:         "Rosethorn Halberd",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(2, 1)},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Rosethorn Halberd — attach it to target non-Human creature you control", effects.AttachSourceToTarget),
				effects.PermanentYouControl("target non-Human creature you control", effects.Creature(), effects.Not(effects.HasSubtype("Human"))),
			),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{5}")},
	})
}
