package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flashfreeze — Instant {1}{U} (Foundations, #4465):
//
//	"Counter target red or green spell."
//
// Precise Redaction's shape in the other two colours. The colour clause is
// checked as the spell is cast and again as it resolves (CR 601.2c, CR
// 608.2b), so a red-and-white spell is a legal target and a white one is
// not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "eaf98e03-729b-4145-b2af-c910c415c15d",
		Name:         "Flashfreeze",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target red or green spell", effects.Or(effects.OfColor("R"), effects.OfColor("G"))),
		OnResolve:    effects.CounterTheTargetSpell,
	})
}
