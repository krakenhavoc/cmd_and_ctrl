package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Order of the White Shield — Creature — Human Knight {W}{W}, 2/1 (Ice Age, S82
// OLD, #5636):
//
//	"Protection from black
//	 {W}: This creature gains first strike until end of turn.
//	 {W}{W}: This creature gets +1/+0 until end of turn."
//
// Knight of Stromgald's shape in white.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e0b3cfc1-68d4-46a8-9ae4-71bdb9c87889",
		Name:            "Order of the White Shield",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"protection from black"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}: This creature gains first strike until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{W}"),
			Effect:  effects.ThisGainsKeywordUntilEndOfTurn("first strike", "Order of the White Shield — first strike until end of turn"),
		}, {
			Label:   "{W}{W}: This creature gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1}},
			Cost:    effects.ManaCost("{W}{W}"),
			Effect:  effects.ThisGetsUntilEndOfTurn(1, 0, "Order of the White Shield — +1/+0 until end of turn"),
		}},
	})
}
