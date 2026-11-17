package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "a7ee67a6-0ac7-4079-a947-4e52d09ace8e",
		Name:         "Axiom Engraver",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{b10EntersWithCounters(oilCounter, 2, "Axiom Engraver: enters with two oil counters on it")},
		Activated: []ActivatedAbility{{
			Label:   "{T}, Remove an oil counter from this creature, Discard a card: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(TapCost(), RemoveCountersFromThis(oilCounter, 1), DiscardACard()),
			Effect:  ebDrawACard,
		}},
	})
}
