package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stone Rain — Sorcery {2}{R} (Limited Edition Alpha, #5642):
//
//	"Destroy target land."
//
// One target clause and the shared "Destroy target [permanent]." body
// (Sinkhole and Ice Storm are the same sentence in black and green).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6e880df6-fc00-43d2-a9c8-f575f40b78c6",
		Name:         "Stone Rain",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target land", effects.Land()),
		OnResolve:    effects.DestroyTheTargetPermanent,
	})
}
