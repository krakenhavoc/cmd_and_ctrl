package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Insight Engine — Artifact {2}{U} (EDHREC rank 2478):
//
//	"{2}, {T}: Put a charge counter on this artifact, then draw a card
//	 for each charge counter on it."
//
// A card-draw rock that grows: one card the first turn, two the
// next, three after that. A CR 602 activation with a mana and a tap
// component — an artifact, so no summoning sickness — whose effect
// puts the charge counter on through AddCounter (a counter doubler
// applies, and the draw counts the doubled total, as printed) and
// then draws the live count. An Engine that left the battlefield in
// response draws its last-known count rather than nothing (CR
// 113.7a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "83409a22-10de-4d51-90a3-e0579ca8cbea",
		Name:         "Insight Engine",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}: Put a charge counter on Insight Engine, then draw a card for each charge counter on it.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Effect:  effects.B23ChargeThenDrawPerCharge,
		}},
	})
}
