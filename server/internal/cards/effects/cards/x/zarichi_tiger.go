package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Zarichi Tiger — Creature — Cat {3}{W}, 2/3 (Gatecrash, #5565):
//
//	"{1}{W}, {T}: You gain 2 life."
//
// Ghost-Lit Redeemer's first ability at {1}{W}. The tap cost waits out
// summoning sickness (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0749e79f-e61d-4037-bb39-75f1ebf46c5d",
		Name:         "Zarichi Tiger",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{W}, {T}: You gain 2 life.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{1}{W}"), effects.TapCost()),
			Effect:  effects.Do(effects.GainLife{Amount: 2}),
		}},
	})
}
