package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kessig Wolfrider — Creature — Human Knight {R}, 1/2 (Innistrad: Crimson
// Vow, #5495):
//
//	"Menace
//	 {2}{R}, {T}, Exile three cards from your graveyard: Create a 3/2 red
//	 Wolf creature token."
//
// Menace rides PrintedKeywords. The exile is a cost component the
// activator names at announce (CR 602.2b); the ability can't be activated
// with fewer than three cards in the graveyard. The token is the printed
// 3/2 red Wolf of Crimson Vow's token card.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d9c42e56-1818-4b73-9cd6-1c5aaf6a5b56",
		Name:            "Kessig Wolfrider",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{R}, {T}, Exile three cards from your graveyard: Create a 3/2 red Wolf creature token.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost: effects.Plus(effects.ManaCost("{2}{R}"), effects.TapCost(),
				effects.ExileFromGraveyard(3, "three cards from your graveyard", nil)),
			Effect: effects.Do(effects.CreateToken{Template: effects.TokenCard("3/2 red Wolf"), N: 1}),
		}},
	})
}
