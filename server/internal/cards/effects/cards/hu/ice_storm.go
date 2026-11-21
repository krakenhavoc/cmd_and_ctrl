package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ice Storm — Sorcery {2}{G} (Limited Edition Alpha, #5642):
//
//	"Destroy target land."
//
// Stone Rain's sentence (stone_rain.go): one target clause and the
// shared "Destroy target [permanent]." body.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a0b97e33-2d0c-4800-a9f5-9cd9be651ea8",
		Name:         "Ice Storm",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target land", effects.Land()),
		OnResolve:    effects.DestroyTheTargetPermanent,
	})
}
