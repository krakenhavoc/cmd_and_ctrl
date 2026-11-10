package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mindless Automaton — Artifact Creature — Construct {4}, 0/0
// (Foundations Commander):
//
//	"This creature enters with two +1/+1 counters on it.
//	 {1}, Discard a card: Put a +1/+1 counter on this creature.
//	 Remove two +1/+1 counters from this creature: Draw a card."
//
// Afiya Grove's entry counters, a discard cost and a counter-removal cost
// (paid at announce, so a response cannot spend the same counters twice).
// The 0/0 body dies to state-based actions the moment the second ability
// removes its last counters, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e7441107-57b1-4e88-a9d8-3eb7c338ef07",
		Name:         "Mindless Automaton",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters(game.CounterPlusOne, 2, "Mindless Automaton: enters with two +1/+1 counters on it"),
		},
		Activated: []ActivatedAbility{
			{
				Label:   "{1}, Discard a card: Put a +1/+1 counter on this creature.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    Plus(ManaCost("{1}"), DiscardACard()),
				Effect:  Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1}),
			},
			{
				Label:   "Remove two +1/+1 counters from this creature: Draw a card.",
				Purpose: game.Purpose{Draws: 1, Answers: game.AnswerValue},
				Cost:    RemoveCountersFromThis(game.CounterPlusOne, 2),
				Effect:  Do(DrawCards{N: 1}),
			},
		},
	})
}
