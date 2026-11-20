package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Zarichi Tiger — Creature — Cat {3}{W}, 2/3 (Gatecrash, #5565):
//
//	"{1}{W}, {T}: You gain 2 life."
//
// Ghost-Lit Redeemer's first ability at {1}{W}. The tap cost waits out
// summoning sickness (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0749e79f-e61d-4037-bb39-75f1ebf46c5d",
		Name:         "Zarichi Tiger",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}{W}, {T}: You gain 2 life.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{1}{W}"), TapCost()),
			Effect:  Do(GainLife{Amount: 2}),
		}},
	})
}
