package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kjeldoran Knight — Creature — Human Knight {W}{W}, 1/1 (Ice Age, S82 OLD batch 849, #5636):
//
//	"Banding"
//	"{1}{W}: This creature gets +1/+0 until end of turn."
//	"{W}{W}: This creature gets +0/+2 until end of turn."
//
// Keywords ride PrintedKeywords; banding is the engine's (game/banding.go, ADR 0224).
// Each pump is a mana-only, untapped, repeatable activation (Bellows Lizard's shape).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5d7747b4-061d-4459-ba80-51361526cbfa",
		Name:            "Kjeldoran Knight",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordBanding},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{1}{W}: This creature gets +1/+0 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1}},
				Cost:    effects.ManaCost("{1}{W}"),
				Effect:  effects.ThisGetsUntilEndOfTurn(1, 0, "Kjeldoran Knight — +1/+0 until end of turn"),
			},
			{
				Label:   "{W}{W}: This creature gets +0/+2 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Toughness: 2}},
				Cost:    effects.ManaCost("{W}{W}"),
				Effect:  effects.ThisGetsUntilEndOfTurn(0, 2, "Kjeldoran Knight — +0/+2 until end of turn"),
			},
		},
	})
}
