package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lathnu Sailback — Creature — Lizard {4}{R}, 5/4 (Aether Revolt, #5536):
//
//	(no rules text)
//
// A vanilla body, registered so the catalog knows it (see bear_cub.go).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9c30147a-1131-402e-bdaf-9006b1022b83",
		Name:         "Lathnu Sailback",
		Completeness: effects.CompletenessFull,
	})
}
