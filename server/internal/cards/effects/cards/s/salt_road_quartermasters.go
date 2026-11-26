package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Salt Road Quartermasters — Creature — Human Soldier {2}{G}, 1/1 (Dragons of Tarkir):
//
//	"This creature enters with two +1/+1 counters on it.
//	 {2}{G}, Remove a +1/+1 counter from this creature: Put a +1/+1 counter on
//	 target creature."
//
// Removing the counter is a cost (CR 602.2b), so the ability needs a counter to
// activate and a target on the announcement.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9398fbb4-a2b9-4ae4-9e0f-2553eb570501",
		Name:         "Salt Road Quartermasters",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.B10EntersWithCounters(game.CounterPlusOne, 2, "Salt Road Quartermasters enters with two +1/+1 counters on it"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{G}, Remove a +1/+1 counter from this creature: Put a +1/+1 counter on target creature.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.Plus(effects.ManaCost("{2}{G}"), effects.RemoveCountersFromThis(game.CounterPlusOne, 1)),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.PutPlusOneCounterOnEachLegalTarget,
		}},
	})
}
