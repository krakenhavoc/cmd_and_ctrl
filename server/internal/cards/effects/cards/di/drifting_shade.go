package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Drifting Shade — Creature — Shade {3}{B}, 1/1 (Magic 2012, #5573):
//
//	"Flying
//	 {B}: This creature gets +1/+1 until end of turn."
//
// Dungeon Shade's text: flying, and a {B} pump on itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fa3883c7-4fb4-48b9-847b-9a94d1988dd7",
		Name:            "Drifting Shade",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}: This creature gets +1/+1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1, Toughness: 1}},
			Cost:    effects.ManaCost("{B}"),
			Effect:  effects.ThisGetsUntilEndOfTurn(1, 1, "Drifting Shade — +1/+1 until end of turn"),
		}},
	})
}
