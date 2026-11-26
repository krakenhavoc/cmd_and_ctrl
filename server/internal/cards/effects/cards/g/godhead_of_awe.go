package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Godhead of Awe — {W/U}{W/U}{W/U}{W/U}{W/U} Creature — Spirit Avatar 4/4 (Shadowmoor,
// S82 older sets):
//
//	"Flying
//	 Other creatures have base power and toughness 1/1."
//
// Kudo, King Among Bears' base-P/T set (layer 7b) with 1/1 over the other
// creatures on both sides of the table. A later +N/+N effect or counter still
// applies on top (CR 613.4c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1ed5751a-4b45-4cd2-b19c-ac84508de4f8",
		Name:            "Godhead of Awe",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{{
			Label:     "Other creatures have base power and toughness 1/1.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7B_Set,
			AppliesTo: effects.GapB52OtherCreatures,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power = 1
				c.Toughness = 1
			},
		}},
	})
}
