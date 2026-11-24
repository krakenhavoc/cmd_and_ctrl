package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Squirrel Wrangler — Creature — Human Druid {2}{G}{G}, 2/2 (Secret Lair
// Drop):
//
//	"{1}{G}, Sacrifice a land: Create two 1/1 green Squirrel creature tokens.
//	 {1}{G}, Sacrifice a land: Squirrel creatures get +1/+1 until end of turn."
//
// Two activated abilities, each with the sacrifice of a land as a cost. The
// tokens are the shared 1/1 green Squirrel. The pump gives +1/+1 to every
// Squirrel creature on the battlefield as it resolves, whoever controls it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4021b6b4-9b87-4dd5-a5df-79540bedbe87",
		Name:         "Squirrel Wrangler",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{1}{G}, Sacrifice a land: Create two 1/1 green Squirrel creature tokens.",
				Purpose: game.Purpose{Answers: game.AnswerSacOutlet | game.AnswerMakesBlocker, Tokens: 2, CreatureTokens: 2},
				Cost:    effects.Plus(effects.ManaCost("{1}{G}"), effects.SacrificeN(1, "a land", effects.Land())),
				Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 green Squirrel"), N: 2}),
			},
			{
				Label:   "{1}{G}, Sacrifice a land: Squirrel creatures get +1/+1 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerSacOutlet | game.AnswerPump},
				Cost:    effects.Plus(effects.ManaCost("{1}{G}"), effects.SacrificeN(1, "a land", effects.Land())),
				Effect:  effects.S82OldB024SquirrelsGetPlusOne,
			},
		},
	})
}
