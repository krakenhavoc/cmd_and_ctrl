package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Jund Battlemage — Creature — Human Shaman {2}{R}, 2/2 (Shards of Alara, S82
// OLD batch 512, #5588):
//
//	"{B}, {T}: Target player loses 1 life.
//	 {G}, {T}: Create a 1/1 green Saproling creature token."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d2567448-f8a5-4bf3-9804-f5aac3b645ce",
		Name:         "Jund Battlemage",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{B}, {T}: Target player loses 1 life.",
				Cost:    effects.Plus(effects.ManaCost("{B}"), effects.TapCost()),
				Targets: effects.TargetPlayer("target player"),
				Effect:  effects.S82OldB512TargetPlayerLosesOne,
			},
			{
				Label:   "{G}, {T}: Create a 1/1 green Saproling creature token.",
				Purpose: game.Purpose{Tokens: 1, Answers: game.AnswerMakesBlocker},
				Cost:    effects.Plus(effects.ManaCost("{G}"), effects.TapCost()),
				Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 green Saproling"), N: 1}),
			},
		},
	})
}
