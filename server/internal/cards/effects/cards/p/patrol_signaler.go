package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Patrol Signaler — {1}{W} Creature — Kithkin Soldier, 1/1:
//
//	"{1}{W}, {Q}: Create a 1/1 white Kithkin Soldier creature token. ({Q} is the
//	 untap symbol.)"
func init() {
	cost := effects.ManaCost("{1}{W}")
	cost.UntapSelf = true
	effects.Register(effects.Spec{
		OracleID:     "1a8d7714-a566-4195-8605-550929584c61",
		Name:         "Patrol Signaler",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{W}, {Q}: Create a 1/1 white Kithkin Soldier creature token.",
			Cost:    cost,
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker, Tokens: 1, CreatureTokens: 1},
			Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Kithkin Soldier"), N: 1}),
		}},
	})
}
