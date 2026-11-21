package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Deity of Scars — Creature — Spirit Avatar {B/G}{B/G}{B/G}{B/G}{B/G}, 7/7 (Eventide, #5590):
//
//	"Trample
//	 This creature enters with two -1/-1 counters on it.
//	 {B/G}, Remove a -1/-1 counter from this creature: Regenerate this
//	 creature."
//
// Gnarlbark Elm's entry counters, and a regeneration row whose cost removes
// one of them (a -1/-1 counter only, as printed).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "bf7c1fa8-0460-4be5-b00d-72c0960c4149",
		Name:            "Deity of Scars",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Replacements: []game.ReplacementEffect{
			effects.B10EntersWithCounters(game.CounterMinusOne, 2, "Deity of Scars: enters with two -1/-1 counters"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{B/G}, Remove a -1/-1 counter from this creature: Regenerate this creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.Plus(effects.ManaCost("{B/G}"), effects.RemoveCountersFromThis(game.CounterMinusOne, 1)),
			Effect:  effects.RegenerateThisSource,
		}},
	})
}
