package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Masked Blackguard — Creature — Human Rogue {1}{B}, 2/1 (Ravnica: Clue
// Edition, Murders at Karlov Manor):
//
//	"Flash
//	 {2}{B}: This creature gets +1/+1 until end of turn."
//
// Flash is a printed keyword; the pump is the shared self-pump body.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8ccd20c7-b7fc-4ab1-a98b-3bbcc2079339",
		Name:            "Masked Blackguard",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{B}: This creature gets +1/+1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1, Toughness: 1}},
			Cost:    effects.ManaCost("{2}{B}"),
			Effect:  effects.ThisGetsUntilEndOfTurn(1, 1, "Masked Blackguard — +1/+1 until end of turn"),
		}},
	})
}
