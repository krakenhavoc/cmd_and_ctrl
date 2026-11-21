package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rend Spirit — Instant {2}{B} (Champions of Kamigawa, #5603):
//
//	"Destroy target Spirit."
//
// Any permanent with the Spirit subtype, as it is now (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cca698b3-b86f-45a2-aa3f-bab7d5a56e8a",
		Name:         "Rend Spirit",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target Spirit", effects.HasSubtype("Spirit")),
		OnResolve:    effects.DestroyTheTargetPermanent,
	})
}
