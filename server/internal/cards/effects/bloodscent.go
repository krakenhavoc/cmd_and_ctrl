package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bloodscent — Instant, {3}{G}:
//
//	"All creatures able to block target creature this turn do so."
//
// Alluring Scent at instant speed (#1684) — cast after attackers are
// declared, it turns a combat the defender had already planned.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b2aabfaf-5996-4b41-987a-6bd941f6ae88",
		Name:         "Bloodscent",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return lureTargetForTheTurn(ctx, "Bloodscent — all creatures able to block it do so")
		},
	})
}
