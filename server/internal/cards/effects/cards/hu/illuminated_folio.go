package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Illuminated Folio — Artifact — Book {5} (Shadowmoor, #5591):
//
//	"{1}, {T}, Reveal two cards from your hand that share a color: Draw
//	 a card."
//
// The shared-colour reveal is RevealNSharingAColor (#6208): a colourless
// card shares no colour (CR 105.2c), so a hand with no sharing pair cannot
// pay, and revealing moves nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0440f85f-8128-4399-9a3d-51f39e196c96",
		Name:         "Illuminated Folio",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}, Reveal two cards from your hand that share a color: Draw a card.",
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost(), effects.RevealNSharingAColor(2, "two cards from your hand that share a color")),
			Purpose: game.Purpose{Draws: 1, Answers: game.AnswerValue},
			Effect:  effects.DrawOneCardEffect,
		}},
	})
}
