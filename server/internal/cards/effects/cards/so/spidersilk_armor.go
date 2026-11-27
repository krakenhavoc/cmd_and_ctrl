package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Spidersilk Armor — Enchantment {2}{G} (Mercadian Masques, S82 OLD, #5619):
//
//	"Creatures you control get +0/+1 and have reach."
//
// A layer-7c modification and a layer-6 keyword grant over the creatures its
// controller controls, the Armor's own timestamp for both (Balthier and Fran's
// pair for Vehicles). Creatures that arrive later are covered, since the
// statics are re-read on every pass.
//
// No simplification.
func init() {
	creaturesYouControl := func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.IsCreature() && target.Controller == source.Controller
	}
	effects.Register(effects.Spec{
		OracleID:     "ad73e5eb-197c-45a3-9d20-a406ba126c79",
		Name:         "Spidersilk Armor",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			{
				Label:     "Creatures you control get +0/+1.",
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7C_Modify,
				AppliesTo: creaturesYouControl,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Toughness++
				},
			},
			effects.KeywordGrant(creaturesYouControl, "reach"),
		},
	})
}
