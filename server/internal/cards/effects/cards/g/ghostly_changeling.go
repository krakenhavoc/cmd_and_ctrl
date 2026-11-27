package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ghostly Changeling — Creature — Shapeshifter {2}{B}, 2/2 (Lorwyn, S82 OLD
// b548, #5593):
//
//	"Changeling (This card is every creature type.)
//	 {1}{B}: This creature gets +1/+1 until end of turn."
//
// Changeling is the printed keyword; the ability is firebreathing through
// ThisGetsUntilEndOfTurn, with no {T}, so it may be activated any number of
// times.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4e779540-e667-457e-9071-4cbbd777d9a9",
		Name:            "Ghostly Changeling",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordChangeling},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}: This creature gets +1/+1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1, Toughness: 1}},
			Cost:    effects.ManaCost("{1}{B}"),
			Effect:  effects.ThisGetsUntilEndOfTurn(1, 1, "Ghostly Changeling — +1/+1 until end of turn"),
		}},
	})
}
