package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Charge — Instant {W} (Dominaria, #5530):
//
//	"Creatures you control get +1/+1 until end of turn."
//
// Warrior's Honor's team pump: the creatures you control as it resolves
// get +1/+1 until end of turn (CR 611.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "36b5022d-fa62-46ea-97ab-23c0f60f62a5",
		Name:         "Charge",
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.CreaturesYouControlGetUntilEOT(1, 1, "Charge — creatures you control get +1/+1"),
	})
}
