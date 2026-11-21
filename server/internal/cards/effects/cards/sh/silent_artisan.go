package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Silent Artisan — Creature — Giant {3}{W}{W}, 3/5 (Theros, #5561):
//
// A vanilla creature: no rules text, so the spec is its identity.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2298eb28-425d-4e87-990b-6bc0ef9bcfda",
		Name:         "Silent Artisan",
		Completeness: effects.CompletenessFull,
	})
}
