package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wall of Kelp — Creature — Plant Wall {U}{U}, 0/3 (Homelands, S82 OLD b821,
// #5635):
//
//	"Defender
//	{U}{U}, {T}: Create a 0/1 blue Plant Wall creature token with defender named
//	Kelp."
//
// A token row that carries its name, as Murf does. No printed Kelp token card
// exists, so the row is listed in knownUnmatchedTokens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "64722c32-975b-43b9-9b4a-9511c75740d9",
		Name:            "Wall of Kelp",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}{U}, {T}: Create a 0/1 blue Plant Wall creature token with defender named Kelp.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Tokens: 1},
			Cost:    effects.Plus(effects.ManaCost("{U}{U}"), effects.TapCost()),
			Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("0/1 blue Plant Wall named Kelp with defender"), N: 1}),
		}},
	})
}
