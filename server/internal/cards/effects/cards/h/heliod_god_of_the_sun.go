package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Heliod, God of the Sun — Legendary Enchantment Creature — God
// {3}{W}, 5/6 (Commander Masters, #5948):
//
//	"Indestructible
//	 As long as your devotion to white is less than five, Heliod isn't a
//	 creature.
//	 Other creatures you control have vigilance.
//	 {2}{W}{W}: Create a 2/1 white Cleric enchantment creature token."
//
// The God clause is the shared devotion static (CR 700.5). The vigilance
// grant reaches the other creatures you control even while Heliod is
// not a creature itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4ea488b2-5ba0-4565-b928-570c8e03b926",
		Name:            "Heliod, God of the Sun",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		Static: []game.StaticAbility{
			effects.GodUnlessDevotion("W"),
			effects.KeywordGrant(effects.OtherCreaturesYouControl, "vigilance"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{W}{W}: Create a 2/1 white Cleric enchantment creature token.",
			Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1, Answers: game.AnswerMakesBlocker},
			Cost:    effects.ManaCost("{2}{W}{W}"),
			Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("2/1 white Cleric enchantment creature"), N: 1}),
		}},
	})
}
