package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Justice Strike — Instant {R}{W} (Guilds of Ravnica, #5525):
//
//	"Target creature deals damage to itself equal to its power."
//
// The creature is its own damage source (so lifelink, deathtouch and
// protection read it), and the power is read as the spell resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bc63e7ca-e62a-4026-803d-eecdbbedf3c6",
		Name:         "Justice Strike",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.B583TargetCreatureDamagesItself,
	})
}
