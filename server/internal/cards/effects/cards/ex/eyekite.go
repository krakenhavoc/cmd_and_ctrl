package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Eyekite — Creature — Drake {1}{U}, 1/2 (Modern Horizons, #5519):
//
//	"Flying
//	 This creature gets +2/+0 as long as you've drawn two or more cards
//	 this turn."
//
// Messenger Hawk's static: layer 7c, read off this turn's draw tally,
// which is not on the battlefield, so the static declares
// DependsOnHandSize.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fb3f2677-a31d-49ad-a9d1-7324733938d6",
		Name:            "Eyekite",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{{
			Label:     "Gets +2/+0 as long as you've drawn two or more cards this turn.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: effects.S82BroB07DrewTwoOrMore,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power += 2
			},
			DependsOnHandSize: true,
		}},
	})
}
