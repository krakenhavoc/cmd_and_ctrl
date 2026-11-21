package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jenara, Asura of War — Legendary Creature — Angel {G}{W}{U}, 3/3:
//
//	"Flying
//	 {1}{W}: Put a +1/+1 counter on Jenara."
//
// The counter goes on Jenara as the ability resolves, and nothing happens
// if she has left the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d06cd670-7ffc-4295-97d1-0eff042fb6d5",
		Name:            "Jenara, Asura of War",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{W}: Put a +1/+1 counter on Jenara.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{1}{W}"),
			Effect:  effects.PutCounterOnSelf,
		}},
	})
}
