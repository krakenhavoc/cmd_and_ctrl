package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Order of the Ebon Hand — Creature — Cleric Knight {B}{B}, 2/1
// (Fallen Empires, #5637):
//
//	"Protection from white
//	 {B}: This creature gains first strike until end of turn.
//	 {B}{B}: This creature gets +1/+0 until end of turn."
//
// Order of Leitbur's mirror in black: the protection rides
// PrintedKeywords, the two rows change the creature itself until end of
// turn, and neither has {T} in its cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2f65f16f-c6c2-48f0-b556-63cf7d162000",
		Name:            "Order of the Ebon Hand",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"protection from white"},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{B}: This creature gains first strike until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    effects.ManaCost("{B}"),
				Effect:  effects.ThisGainsKeywordUntilEndOfTurn("first strike", "Order of the Ebon Hand — first strike until end of turn"),
			},
			{
				Label:   "{B}{B}: This creature gets +1/+0 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    effects.ManaCost("{B}{B}"),
				Effect:  effects.ThisGetsUntilEndOfTurn(1, 0, "Order of the Ebon Hand — +1/+0 until end of turn"),
			},
		},
	})
}
