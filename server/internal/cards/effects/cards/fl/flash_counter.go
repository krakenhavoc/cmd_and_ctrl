package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flash Counter — Instant {1}{U} (Legends, #5639):
//
//	"Counter target instant spell."
//
// Dispel's clause, on the shared counterTheTargetSpell body: only an
// instant spell can be picked, and the clause is checked again as it
// resolves (CR 608.2b). A countered spell goes to its owner's
// graveyard (CR 701.6a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "75b2dd64-a2f5-4032-a19f-5b485dc44b16",
		Name:         "Flash Counter",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target instant spell", effects.Instant()),
		OnResolve:    effects.CounterTheTargetSpell,
	})
}
