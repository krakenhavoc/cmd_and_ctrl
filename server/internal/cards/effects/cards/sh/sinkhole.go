package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sinkhole — Sorcery {B}{B} (Limited Edition Alpha, #5642):
//
//	"Destroy target land."
//
// Stone Rain's sentence (stone_rain.go): one target clause and the
// shared "Destroy target [permanent]." body.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5a46ad2a-35b1-4dd5-b7c3-fec36b7c67ab",
		Name:         "Sinkhole",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target land", effects.Land()),
		OnResolve:    effects.DestroyTheTargetPermanent,
	})
}
