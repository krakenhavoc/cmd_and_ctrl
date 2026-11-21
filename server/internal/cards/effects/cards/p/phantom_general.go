package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Phantom General — Creature — Spirit Soldier {3}{W}, 2/3:
//
//	"Creature tokens you control get +1/+1."
//
// A layer 7c modify over creature tokens its controller controls, the same
// applies-to Intangible Virtue uses (b05CreatureTokenYouControl). It is
// re-read on every recompute, so a token made after the General is pumped as
// it enters. The General is not a token and gets nothing from itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b800c3e5-4ed9-4603-a987-d830e90cf25e",
		Name:         "Phantom General",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "Creature tokens you control get +1/+1.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: effects.B05CreatureTokenYouControl,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power++
				c.Toughness++
			},
		}},
	})
}
