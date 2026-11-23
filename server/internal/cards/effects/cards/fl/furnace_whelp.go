package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Furnace Whelp — Creature — Dragon {2}{R}{R}, 2/2 (Jumpstart 2022, #5948):
//
//	"Flying
//	 {R}: This creature gets +1/+0 until end of turn."
//
// Firebreathing through the shared ThisGetsUntilEndOfTurn body (Bellows
// Lizard, Shivan Dragon's shape): no {T}, so it can be activated any number of
// times, summoning sickness or not. Flying is a printed keyword.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8422f1e0-00ca-4ffb-a6b2-e2c9f96d7f23",
		Name:            "Furnace Whelp",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}: This creature gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1}},
			Cost:    effects.ManaCost("{R}"),
			Effect:  effects.ThisGetsUntilEndOfTurn(1, 0, "Furnace Whelp — +1/+0 until end of turn"),
		}},
	})
}
