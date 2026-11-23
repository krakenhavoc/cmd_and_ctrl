package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mad Prophet — {3}{R} Creature — Human Shaman 2/2 (Avacyn Restored, #5570):
//
//	"Haste
//	 {T}, Discard a card: Draw a card."
//
// Charging Strifeknight's haste and rummage row.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "91081181-aec6-47de-b8ef-241a2f4fe880",
		Name:            "Mad Prophet",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Discard a card: Draw a card.",
			Cost:    effects.Plus(effects.TapCost(), effects.DiscardACard()),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect:  effects.Do(effects.DrawCards{N: 1}),
		}},
	})
}
