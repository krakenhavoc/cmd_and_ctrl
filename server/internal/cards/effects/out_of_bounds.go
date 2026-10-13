package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Out of Bounds — Instant {3}{U}:
//
//	"Assist (Another player can pay up to {3} of this spell's cost.)
//	 Counter target spell."
//
// Cancel's shape with the assist keyword and a different cost. The
// counter is CounterTarget on the picked spell.
func init() {
	Register(Spec{
		OracleID:        "431b2e52-1290-47b1-920d-2c1817553d96",
		Name:            "Out of Bounds",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordAssist},
		Targets:         TargetSpell("target spell"),
		OnResolve:       counterTheTargetSpell,
	})
}
