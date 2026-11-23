package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Leonin Snarecaster — Creature — Cat Soldier {1}{W}, 2/1 (Jumpstart 2022):
//
//	"When this creature enters, you may tap target creature."
//
// The target is chosen as the trigger goes on the stack and the "you may" is
// asked as it resolves (Elite Skirmisher's shape). A target gone by
// resolution is not tapped (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "75f269df-ee49-44db-b256-4d005e3022fb",
		Name:         "Leonin Snarecaster",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.Optional(
					effects.WhenThisEnters("Leonin Snarecaster — you may tap target creature", effects.TapTheFirstLegalTarget),
					"Leonin Snarecaster — tap target creature?"),
				effects.TargetCreature("target creature"),
			),
		},
	})
}
