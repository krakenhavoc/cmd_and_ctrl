package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mad Prophet — {3}{R} Creature — Human Shaman 2/2 (Avacyn Restored, #5570):
//
//	"Haste
//	 {T}, Discard a card: Draw a card."
//
// Charging Strifeknight's haste and rummage row.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "91081181-aec6-47de-b8ef-241a2f4fe880",
		Name:            "Mad Prophet",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Activated: []ActivatedAbility{{
			Label:   "{T}, Discard a card: Draw a card.",
			Cost:    Plus(TapCost(), DiscardACard()),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect:  Do(DrawCards{N: 1}),
		}},
	})
}
