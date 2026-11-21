package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Axiom Engraver — Creature — Phyrexian Wizard {1}{R}, 1/3 (Phyrexia: All
// Will Be One):
//
//	"This creature enters with two oil counters on it.
//	 {T}, Remove an oil counter from this creature, Discard a card: Draw a
//	 card."
//
// Tap, the oil counter and the discard are all costs paid at announcement.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a7ee67a6-0ac7-4079-a947-4e52d09ace8e",
		Name:         "Axiom Engraver",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.B10EntersWithCounters(effects.OilCounter, 2, "Axiom Engraver: enters with two oil counters on it")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Remove an oil counter from this creature, Discard a card: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.TapCost(), effects.RemoveCountersFromThis(effects.OilCounter, 1), effects.DiscardACard()),
			Effect:  effects.EbDrawACard,
		}},
	})
}
