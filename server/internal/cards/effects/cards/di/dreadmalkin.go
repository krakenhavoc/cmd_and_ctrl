package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dreadmalkin — Creature — Zombie Cat {B}, 1/1 (Ravnica Remastered, #5948):
//
//	"Menace
//	 {2}{B}, Sacrifice another creature or planeswalker: Put two +1/+1
//	 counters on this creature."
//
// "Another" is object identity (SacrificeAnotherN).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "118cf682-a532-42e9-8d91-6ce6f0d31ace",
		Name:            "Dreadmalkin",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{B}, Sacrifice another creature or planeswalker: Put two +1/+1 counters on this creature.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet | game.AnswerPump},
			Cost:    effects.Plus(effects.ManaCost("{2}{B}"), effects.SacrificeAnotherN(1, "another creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker()))),
			Effect:  effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 2}),
		}},
	})
}
