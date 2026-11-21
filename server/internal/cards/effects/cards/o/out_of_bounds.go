package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Out of Bounds — Instant {3}{U}:
//
//	"Assist (Another player can pay up to {3} of this spell's cost.)
//	 Counter target spell."
//
// Cancel's shape with the assist keyword and a different cost. The
// counter is CounterTarget on the picked spell.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "431b2e52-1290-47b1-920d-2c1817553d96",
		Name:            "Out of Bounds",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordAssist},
		Targets:         effects.TargetSpell("target spell"),
		OnResolve:       effects.CounterTheTargetSpell,
	})
}
