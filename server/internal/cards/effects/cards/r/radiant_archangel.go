package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Radiant, Archangel — Legendary Creature — Angel {3}{W}{W}, 3/3 (Urza's Legacy, S82 OLD b715, #5623):
//
//	"Flying, vigilance
//	Radiant gets +1/+1 for each other creature on the battlefield with flying."
//
// A layer 7c static that counts every OTHER creature with flying on the
// battlefield, whoever controls it, through the layered view and read live
// (RadiantOtherFlyers), so a flyer arriving or losing flying changes it at
// once. Radiant itself is never counted.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "859cbe11-e1e6-4309-94da-8725db1b3fc2",
		Name:            "Radiant, Archangel",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Static: []game.StaticAbility{{
			Label:     "Radiant gets +1/+1 for each other creature on the battlefield with flying.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, self *game.Card, g *game.Game, _ *game.Card) {
				n := effects.RadiantOtherFlyers(g, self)
				c.Power += n
				c.Toughness += n
			},
		}},
	})
}
