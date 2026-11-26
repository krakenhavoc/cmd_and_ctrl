package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jhessian Lookout — {1}{U} Creature — Human Scout 2/1 (Shards of Alara, S82 OLD b513):
//
//	(no rules text)
//
// A vanilla body, registered so the catalog knows it (see grizzly_bears.go).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "acc07d0d-141b-4169-9643-6f20739e91be",
		Name:         "Jhessian Lookout",
		Completeness: effects.CompletenessFull,
	})
}
