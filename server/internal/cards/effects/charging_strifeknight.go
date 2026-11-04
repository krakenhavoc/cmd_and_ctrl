package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Charging Strifeknight — Creature — Spirit Knight {2}{R}, 3/3, Secrets of
// Strixhaven (#4456):
//
//	"Haste
//	 {T}, Discard a card: Draw a card."
//
// Haste is a printed keyword. The discard is a cost the activator picks from
// their hand (CR 601.2h via 602.2b), so the draw only happens when a card was
// really discarded.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b0b9a7c1-f515-4627-a050-b595d6641611",
		Name:            "Charging Strifeknight",
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
