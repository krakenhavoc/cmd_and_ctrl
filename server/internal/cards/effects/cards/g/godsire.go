package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Godsire — Creature — Beast {4}{R}{G}{G}{W}, 8/8 (Shards of Alara, S82 OLD
// batch 510, #5588):
//
//	"Vigilance
//	 {T}: Create an 8/8 Beast creature token that's red, green, and white."
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4252d7f4-c582-4a21-b3af-829c84cb0b09",
		Name:            "Godsire",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Create an 8/8 Beast creature token that's red, green, and white.",
			Purpose: game.Purpose{Tokens: 1, Answers: game.AnswerMakesBlocker},
			Cost:    effects.TapCost(),
			Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("8/8 red and green and white Beast"), N: 1}),
		}},
	})
}
