package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Reaver Ambush — Instant {2}{B} (Rivals of Ixalan, #5531):
//
//	"Exile target creature with power 3 or less."
//
// Epic Downfall's shape with a power bound: the creature's power is read
// through the layers as the spell is cast and again as it resolves
// (CR 608.2b), so a creature pumped past 3 in response is no longer a
// legal target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0385be55-8e8e-4e7d-b6b7-51adc6eb98fa",
		Name:         "Reaver Ambush",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature with power 3 or less", effects.PowerLE(3)),
		OnResolve:    effects.ResolveSteps(effects.Instruction(effects.ExileTheFirstTargetIfLegal)),
	})
}
