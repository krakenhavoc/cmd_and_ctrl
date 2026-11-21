package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Outnumber — Instant {R} (Jumpstart 2022, #5948):
//
//	"Outnumber deals damage to target creature equal to the number of
//	 creatures you control."
//
// The count is taken as the spell resolves, through the layers, so a
// Vehicle crewed in response counts and a creature that died does not. A
// target that became illegal leaves the spell with no effect (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b72efd67-a234-44af-8296-3e182292c7e6",
		Name:         "Outnumber",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.GapB171Outnumber,
	})
}
