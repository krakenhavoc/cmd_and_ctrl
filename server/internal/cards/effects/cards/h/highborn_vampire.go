package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Highborn Vampire — Creature — Vampire Warrior {3}{B}, 4/3 (Foundations, #4465):
//
//	(no rules text)
//
// A vanilla body, registered so the catalog knows it (see ordinary_bear.go).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1adcf587-9035-4299-a2bd-62f805fdd897",
		Name:         "Highborn Vampire",
		Completeness: effects.CompletenessFull,
	})
}
