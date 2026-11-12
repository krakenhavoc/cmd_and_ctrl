package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Clockwork Vorrac — Artifact Creature — Boar Beast {5}, 0/0 (Mirrodin,
// #5606):
//
//	"Trample
//	 This creature enters with four +1/+1 counters on it.
//	 Whenever this creature attacks or blocks, remove a +1/+1 counter
//	 from it at end of combat.
//	 {T}: Put a +1/+1 counter on this creature."
//
// The Clockwork trigger (clockwork.go) and a rewind that taps it, so it
// winds up only while it is neither attacking without vigilance nor
// summoning sick (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e1a64bba-eaff-4f59-b85d-7df0b13e2e6b",
		Name:            "Clockwork Vorrac",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters(game.CounterPlusOne, 4, "Clockwork Vorrac: enters with four +1/+1 counters"),
		},
		Triggered: []game.TriggeredAbility{
			clockworkLosesACounterAtEndOfCombat("Clockwork Vorrac"),
		},
		Activated: []ActivatedAbility{{
			Label:   "{T}: Put a +1/+1 counter on this creature.",
			Purpose: game.Purpose{Answers: game.AnswerPump, SourceCounters: 1},
			Cost:    TapCost(),
			Effect:  putACounterOnThis(game.CounterPlusOne),
		}},
	})
}
