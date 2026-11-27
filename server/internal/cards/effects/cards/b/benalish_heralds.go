package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Benalish Heralds — Creature — Human Soldier {3}{W}, 2/4 (Invasion, S82 #5616):
//
//	"{3}{U}, {T}: Draw a card."
//
// A plain mana-and-tap activation of one draw.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "09ee1332-741f-4c55-abc3-8bcff9031cd9",
		Name:         "Benalish Heralds",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{U}, {T}: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
			Cost:    effects.Plus(effects.ManaCost("{3}{U}"), effects.TapCost()),
			Effect:  effects.Do(effects.DrawCards{N: 1}),
		}},
	})
}
