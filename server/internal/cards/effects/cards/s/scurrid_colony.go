package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scurrid Colony — Creature — Squirrel {1}{G}, 2/2 (Strixhaven: School of
// Mages, #5502):
//
//	"Reach
//	 This creature gets +2/+2 as long as you control eight or more lands."
//
// Reach is the printed keyword; the pump is Woodborn Behemoth's layer-7c
// static at +2/+2, read off the lands its controller controls whenever the
// layers are applied.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7dd68fd9-f5a7-49ee-aa36-93d646254335",
		Name:            "Scurrid Colony",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Static: []game.StaticAbility{
			{Label: "This creature gets +2/+2 as long as you control eight or more lands.", Layer: game.Layer7PT, SubLayer: game.SubLayer7C_Modify, AppliesTo: effects.WoodbornBehemothEightLands, Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power += 2
				c.Toughness += 2
			}},
		},
	})
}
