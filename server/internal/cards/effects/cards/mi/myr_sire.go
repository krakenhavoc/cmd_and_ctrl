package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Myr Sire — Artifact Creature — Phyrexian Myr {2}, 1/1 (Commander Masters):
//
//	"When this creature dies, create a 1/1 colorless Phyrexian Myr artifact
//	 creature token."
//
// A plain dies trigger over the printed token (Commander Masters' Phyrexian
// Myr, tcmm).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "25794033-bd17-4056-b116-68d3acf52569",
		Name:         "Myr Sire",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Myr Sire — create a 1/1 colorless Phyrexian Myr artifact creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 colorless Phyrexian Myr artifact"), N: 1})),
		},
	})
}
