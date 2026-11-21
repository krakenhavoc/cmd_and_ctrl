package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Precise Redaction — Instant {1}{U}:
//
//	"Counter target white or black spell."
//
// The colour clause is enforced at announce and again at resolution
// (CR 601.2c, CR 608.2b), so a spell that is both white and blue is a
// legal target and a green one is not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c251b676-0e98-4047-bed1-d72c83aa8da0",
		Name:         "Precise Redaction",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target white or black spell", effects.Or(effects.OfColor("W"), effects.OfColor("B"))),
		OnResolve:    effects.CounterTheTargetSpell,
	})
}
