package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Welding Sparks — Instant {2}{R} (Jumpstart 2022, #5948):
//
//	"Welding Sparks deals X damage to target creature, where X is 3 plus the
//	 number of artifacts you control."
//
// The artifacts are counted as the spell resolves, through the layers, so an
// artifact creature or a Vehicle counts and one destroyed in response does
// not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "82db3a7e-ec3f-4f25-996e-9e1c23c86003",
		Name:         "Welding Sparks",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.GapB172WeldingSparks,
	})
}
