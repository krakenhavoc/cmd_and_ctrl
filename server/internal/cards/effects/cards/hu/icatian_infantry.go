package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Icatian Infantry — Creature — Human Soldier {W}, 1/1 (Fallen Empires, S82 OLD
// batch 852, #5637):
//
//	"{1}: This creature gains first strike until end of turn.
//	 {1}: This creature gains banding until end of turn."
//
// Two untapped, repeatable keyword grants (Brightling's shape); banding is the
// engine's (game/banding.go, ADR 0224), and a creature that gains it mid-turn
// bands in the declaration that follows.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9eb274f3-ef55-4545-81b4-29454a48f075",
		Name:         "Icatian Infantry",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{1}: This creature gains first strike until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
				Cost:    effects.ManaCost("{1}"),
				Effect:  effects.ThisGainsKeywordUntilEndOfTurn("first strike", "Icatian Infantry — first strike until end of turn"),
			},
			{
				Label:   "{1}: This creature gains banding until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
				Cost:    effects.ManaCost("{1}"),
				Effect:  effects.ThisGainsKeywordUntilEndOfTurn(game.KeywordBanding, "Icatian Infantry — banding until end of turn"),
			},
		},
	})
}
