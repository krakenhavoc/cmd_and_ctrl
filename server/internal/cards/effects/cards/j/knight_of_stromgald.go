package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Knight of Stromgald — Creature — Human Knight {B}{B}, 2/1 (Ice Age, S82 OLD,
// #5636):
//
//	"Protection from white
//	 {B}: This creature gains first strike until end of turn.
//	 {B}{B}: This creature gets +1/+0 until end of turn."
//
// Black Knight's protection with two self-pump abilities (Battlefield Percher's
// shape).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8cf57287-21b0-4194-8da3-62d00160234f",
		Name:            "Knight of Stromgald",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"protection from white"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}: This creature gains first strike until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{B}"),
			Effect:  effects.ThisGainsKeywordUntilEndOfTurn("first strike", "Knight of Stromgald — first strike until end of turn"),
		}, {
			Label:   "{B}{B}: This creature gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1}},
			Cost:    effects.ManaCost("{B}{B}"),
			Effect:  effects.ThisGetsUntilEndOfTurn(1, 0, "Knight of Stromgald — +1/+0 until end of turn"),
		}},
	})
}
