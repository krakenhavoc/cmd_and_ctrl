package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Carnifex Demon — Creature — Phyrexian Demon {4}{B}{B}, 6/6 (Lorwyn
// Eclipsed Commander, #5368):
//
//	"Flying
//	 This creature enters with two -1/-1 counters on it.
//	 {B}, Remove a -1/-1 counter from this creature: Put a -1/-1 counter
//	 on each other creature."
//
// The removal is a cost paid at announce. "Each other creature" is every
// creature but the Demon, whoever controls it, read once before the first
// counter goes on (cmdr2B25CountersOnEach).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "527e4eec-0983-46a9-a4f7-451080200aea",
		Name:            "Carnifex Demon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters(game.CounterMinusOne, 2, "Carnifex Demon: enters with two -1/-1 counters"),
		},
		Activated: []ActivatedAbility{{
			Label:   "{B}, Remove a -1/-1 counter from this creature: Put a -1/-1 counter on each other creature.",
			Purpose: game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 1}, Answers: game.AnswerRemove},
			Cost:    Plus(ManaCost("{B}"), RemoveCountersFromThis(game.CounterMinusOne, 1)),
			Effect:  cmdr2B25CountersOnEach(true, nil),
		}},
	})
}
