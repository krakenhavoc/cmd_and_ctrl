package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Slagdrill Scrapper — Artifact Creature — Robot Scout {R}, 1/2 (Edge of
// Eternities, #4461):
//
//	"{2}, {T}, Sacrifice another artifact or land: Draw a card."
//
// Grabby Giant's sacrifice-an-artifact-or-land draw with a tap and the
// word "another": the Scrapper is itself an artifact but cannot pay its
// own sacrifice. The {T} needs it free of summoning sickness (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7f1edf4c-46a1-426d-8ce9-757a4c27367d",
		Name:         "Slagdrill Scrapper",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}, Sacrifice another artifact or land: Draw a card.",
			Purpose: game.Purpose{Draws: 1, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost(), effects.SacrificeAnotherN(1, "another artifact or land", effects.Or(effects.Artifact(), effects.Land()))),
			Effect:  effects.Do(effects.DrawCards{N: 1}),
		}},
	})
}
