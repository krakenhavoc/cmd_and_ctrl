package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bastion Enforcer — Creature — Dwarf Soldier {2}{W}, 3/2 (Aether Revolt, #5536):
//
//	(no rules text)
//
// A vanilla body, registered so the catalog knows it (see bear_cub.go).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "87449f64-9500-4967-a073-cde19824e087",
		Name:         "Bastion Enforcer",
		Completeness: effects.CompletenessFull,
	})
}
