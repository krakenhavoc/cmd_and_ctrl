package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Markov Dreadknight — Creature — Vampire Knight {3}{B}{B}, 3/3 (Shadows
// over Innistrad, #5543):
//
//	"Flying
//	 {2}{B}, Discard a card: Put two +1/+1 counters on this creature."
//
// The discard is a cost paid as the ability is activated (CR 602.2b), so
// an empty hand can't activate it. The counters go on only if the source is
// still on the battlefield (Mindless Automaton's shape).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "73e2d8aa-3a23-4458-8cae-e9209a102197",
		Name:            "Markov Dreadknight",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{B}, Discard a card: Put two +1/+1 counters on this creature.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}{B}"), effects.DiscardACard()),
			Effect:  effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 2}),
		}},
	})
}
