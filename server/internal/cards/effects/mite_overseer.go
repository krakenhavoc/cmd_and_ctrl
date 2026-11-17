package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
		return target.IsCreature() && IsToken(*target) &&
			target.Controller == source.Controller && isActivePlayer(g, source.Controller)
	}
	Register(Spec{
		OracleID:        "d92411b3-cb25-4123-a78e-4f785dfdae0c",
		Name:            "Mite Overseer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Purpose:         game.Purpose{Tokens: 1, CreatureTokens: 1},
		Static: []game.StaticAbility{
			{
				Label:     "During your turn, creature tokens you control get +1/+0.",
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7C_Modify,
				AppliesTo: yourTokens,
				Apply:     pumpBy(1, 0),
			},
			KeywordGrant(yourTokens, "first strike"),
		},
		Activated: []ActivatedAbility{{
			Label: "{3}{W/P}: Create a 1/1 colorless Phyrexian Mite artifact creature token with toxic 1 and \"This token can't block.\"",
			Cost:  ManaCost("{3}{W/P}"),
			// Mites can't block, so the token answers nothing at instant speed.
			Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1, Answers: game.AnswerValue},
			Effect:  Do(CreateToken{Template: PhyrexianMiteToken(), N: 1}),
		}},
	})
}
