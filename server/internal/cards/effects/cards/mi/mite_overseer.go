package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mite Overseer — Creature — Phyrexian Soldier {3}{W}, 4/2:
//
//	"First strike
//	 During your turn, creature tokens you control get +1/+0 and have first
//	 strike.
//	 {3}{W/P}: Create a 1/1 colorless Phyrexian Mite artifact creature token
//	 with toxic 1 and 'This token can't block.'"
//
// The pump and the first strike are two statics sharing one window: the
// controller's own turn, re-read on every layer pass.
//
// No simplification.
func init() {
	yourTokens := func(target *game.Card, g *game.Game, source *game.Card) bool {
		return target.IsCreature() && effects.IsToken(*target) &&
			target.Controller == source.Controller && effects.IsActivePlayer(g, source.Controller)
	}
	effects.Register(effects.Spec{
		OracleID:        "d92411b3-cb25-4123-a78e-4f785dfdae0c",
		Name:            "Mite Overseer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Purpose:         game.Purpose{Tokens: 1, CreatureTokens: 1},
		Static: []game.StaticAbility{
			{
				Label:     "During your turn, creature tokens you control get +1/+0.",
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7C_Modify,
				AppliesTo: yourTokens,
				Apply:     effects.PumpBy(1, 0),
			},
			effects.KeywordGrant(yourTokens, "first strike"),
		},
		Activated: []effects.ActivatedAbility{{
			Label: "{3}{W/P}: Create a 1/1 colorless Phyrexian Mite artifact creature token with toxic 1 and \"This token can't block.\"",
			Cost:  effects.ManaCost("{3}{W/P}"),
			// Mites can't block, so the token answers nothing at instant speed.
			Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1, Answers: game.AnswerValue},
			Effect:  effects.Do(effects.CreateToken{Template: effects.PhyrexianMiteToken(), N: 1}),
		}},
	})
}
