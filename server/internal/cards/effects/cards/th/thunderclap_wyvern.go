package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thunderclap Wyvern — Creature — Drake {2}{W}{U}, 2/3 (Magic Origins, #5549):
//
//	"Flash (You may cast this spell any time you could cast an instant.)
//	 Flying
//	 Other creatures you control with flying get +1/+1."
//
// Flash, flying, and Empyrean Eagle's lord over your other flyers.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "15276d3b-a117-44bf-87c3-c17e032e4a26",
		Name:            "Thunderclap Wyvern",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Static: []game.StaticAbility{{
			Label:     "Other creatures you control with flying get +1/+1.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: effects.B29OtherCreaturesYouControlWithFlying,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power++
				c.Toughness++
			},
		}},
	})
}
