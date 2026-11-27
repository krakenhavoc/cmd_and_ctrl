package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Override — Instant {2}{U} (Mirrodin, #5606):
//
//	"Counter target spell unless its controller pays {1} for each artifact you
//	 control."
//
// Spell Rupture's pay-unless with the amount counted as Override resolves
// (CR 608.2h): the artifacts its own controller controls then. With none the
// spell's controller pays nothing and the spell is not countered.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a2845761-7e43-4a88-8fc2-2d95d471b00c",
		Name:         "Override",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target spell"),
		OnResolve:    effects.OldB607OverrideCounter,
	})
}
