package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skycat Sovereign — Creature — Elemental Cat {W}{U}, 1/1 (Starter Commander
// Decks):
//
//	"Flying
//	 This creature gets +1/+1 for each other creature you control with
//	 flying.
//	 {2}{W}{U}: Create a 1/1 white Cat Bird creature token with flying."
//
// The pump is a layer 7c static that counts the other fliers its controller
// controls through the layered view, read live, so a Cat Bird token it makes
// raises it at once. The token is the printed 1/1 white Cat Bird with flying.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "84adce5c-39c7-425e-b163-4a1a3977364b",
		Name:            "Skycat Sovereign",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{{
			Label:     "This creature gets +1/+1 for each other creature you control with flying.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, self *game.Card, g *game.Game, _ *game.Card) {
				n := effects.GapB181OtherFlyersYouControl(g, self)
				c.Power += n
				c.Toughness += n
			},
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{W}{U}: Create a 1/1 white Cat Bird creature token with flying.",
			Cost:    effects.ManaCost("{2}{W}{U}"),
			Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1, Answers: game.AnswerValue},
			Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Cat Bird with flying"), N: 1}),
		}},
	})
}
