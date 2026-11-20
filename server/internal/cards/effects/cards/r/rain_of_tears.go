package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rain of Tears — Sorcery {1}{B}{B} (Portal, #5630):
//
//	"Destroy target land."
//
// Stone Rain's shape (destroyTheTargetPermanent).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "72cecab3-519e-4a23-9623-b423a5c5a251",
		Name:         "Rain of Tears",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target land", effects.Land()),
		OnResolve:    effects.DestroyTheTargetPermanent,
	})
}
