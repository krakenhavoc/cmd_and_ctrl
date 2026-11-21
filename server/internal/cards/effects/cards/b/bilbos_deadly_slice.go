package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bilbo's Deadly Slice — Instant {1}{B}{B} (The Hobbit, HOB):
//
//	"Destroy target creature."
//
// Murder under another name (#4454): same Spec, new oracle id.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4c054431-3667-4e28-9d68-03b43447c35b",
		Name:         "Bilbo's Deadly Slice",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.DestroyTheTargetPermanent,
	})
}
