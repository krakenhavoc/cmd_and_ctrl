package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Unwilling Ingredient — Creature — Frog {B}, 1/1 (Strixhaven: School of
// Mages, #5502):
//
//	"Menace (This creature can't be blocked except by two or more
//	 creatures.)
//	 {2}{B}, Exile this card from your graveyard: You draw a card and you
//	 lose 1 life."
//
// Menace is the printed keyword; the graveyard row is Faerie Dreamthief's,
// activated only from the graveyard with exiling the card as its cost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "40d483fb-0dc7-413e-87f9-8d7cb7883fca",
		Name:            "Unwilling Ingredient",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Activated: []ActivatedAbility{{
			Label:   "{2}{B}, Exile this card from your graveyard: You draw a card and you lose 1 life.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
			Cost:    Plus(ManaCost("{2}{B}"), ExileThis()),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect:  b36DrawAndLoseOne,
		}},
	})
}
