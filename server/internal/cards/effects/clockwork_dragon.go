package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Clockwork Dragon — Artifact Creature — Dragon {7}, 0/0 (Mirrodin,
// #5606):
//
//	"Flying
//	 This creature enters with six +1/+1 counters on it.
//	 Whenever this creature attacks or blocks, remove a +1/+1 counter
//	 from it at end of combat.
//	 {3}: Put a +1/+1 counter on this creature."
//
// The Clockwork trigger (clockwork.go) and a rewind at {3}, at instant
// speed and as often as it is paid for; a Dragon that has left the
// battlefield gets nothing (putACounterOnThis).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3361cf88-171a-4ddd-ac56-a31d42613c80",
		Name:            "Clockwork Dragon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters(game.CounterPlusOne, 6, "Clockwork Dragon: enters with six +1/+1 counters"),
		},
		Triggered: []game.TriggeredAbility{
			clockworkLosesACounterAtEndOfCombat("Clockwork Dragon"),
		},
		Activated: []ActivatedAbility{{
			Label:   "{3}: Put a +1/+1 counter on this creature.",
			Purpose: game.Purpose{Answers: game.AnswerPump, SourceCounters: 1},
			Cost:    ManaCost("{3}"),
			Effect:  putACounterOnThis(game.CounterPlusOne),
		}},
	})
}
