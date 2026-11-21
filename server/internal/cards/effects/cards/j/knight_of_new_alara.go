package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Knight of New Alara — Creature — Human Knight {2}{G}{W}, 2/2
// (Commander deck reprint):
//
//	"Each other multicolored creature you control gets +1/+1 for each of
//	 its colors."
//
// A layer 7c static. Colours are settled in layer 5, before layer 7, so the
// bonus reads the colours the creature has in the characteristic being
// built: a creature that is two colours gets +2/+2, five colours +5/+5, and
// a monocolored or colourless one nothing (it is not multicolored,
// CR 105.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "170856f4-7c9d-494b-9bae-20e4d9eba0da",
		Name:         "Knight of New Alara",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:    "Each other multicolored creature you control gets +1/+1 for each of its colors.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID != source.InstanceID && target.IsCreature() && target.Controller == source.Controller
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				if n := len(c.Colors); n >= 2 {
					c.Power += n
					c.Toughness += n
				}
			},
		}},
	})
}
