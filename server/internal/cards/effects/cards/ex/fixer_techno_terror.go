package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fixer, Techno Terror — Legendary Creature — Human Artificer Villain {1}{B},
// 2/1 (Marvel Super Heroes Commander, #4455):
//
//	"{T}, Pay 2 life: Draw a card. Activate only if an artifact entered the
//	 battlefield under your control this turn."
//
// Erebos's pay-life draw with Master's Manufactory's activation condition:
// the turn's entry tally of artifacts under the activator's control
// (tokens included, one that has since left still counts), checked as the
// ability is activated (CR 602.5b). Tapping and the life are costs, so the
// condition is checked before either is paid.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a47b71df-9357-4118-90c0-f86fca6d362b",
		Name:         "Fixer, Techno Terror",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:     "{T}, Pay 2 life: Draw a card. Activate only if an artifact entered the battlefield under your control this turn.",
			Purpose:   game.Purpose{Answers: game.AnswerValue},
			Cost:      effects.Plus(effects.TapCost(), effects.PayLife(2)),
			Condition: effects.AnArtifactEnteredUnderYouThisTurn,
			Effect:    effects.DrawOneCardEffect,
		}},
	})
}
