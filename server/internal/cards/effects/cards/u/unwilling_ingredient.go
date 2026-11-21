package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "40d483fb-0dc7-413e-87f9-8d7cb7883fca",
		Name:            "Unwilling Ingredient",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{B}, Exile this card from your graveyard: You draw a card and you lose 1 life.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
			Cost:    effects.Plus(effects.ManaCost("{2}{B}"), effects.ExileThis()),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect:  effects.B36DrawAndLoseOne,
		}},
	})
}
