package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Act of Aggression — Instant {3}{R/P}{R/P} (Commander 2015, S82 OLD #5546):
//
//	"({R/P} can be paid with either {R} or 2 life.)
//	 Gain control of target creature an opponent controls until end of turn.
//	 Untap that creature. It gains haste until end of turn."
//
// Act of Treason's three clauses in printed order (s82_old_b303_helpers.go),
// aimed at "a creature an opponent controls". Each Phyrexian symbol of the
// cost is paid with its colour or 2 life (CR 107.4f), as Gut Shot's is.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9d7624ea-ffff-470b-928b-bbac66ddbd1c",
		Name:         "Act of Aggression",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
		OnResolve:    effects.ThreatenTheTarget("Act of Aggression"),
	})
}
