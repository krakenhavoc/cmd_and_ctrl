package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Root Snare — Instant {1}{G} (Core Set 2019, #5527):
//
//	"Prevent all combat damage that would be dealt this turn."
//
// Fog's text: every combat damage event this turn is prevented (CR
// 615.1a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cbf743a5-123f-4747-826d-7f8bb929a52c",
		Name:         "Root Snare",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return PreventAllCombatDamageThisTurn{
				Label: "Root Snare: prevent combat damage",
			}.Apply(ctx)
		},
	})
}
