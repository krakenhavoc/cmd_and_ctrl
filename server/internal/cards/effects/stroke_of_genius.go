package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stroke of Genius — Instant for {X}{2}{U}:
//
//	"Target player draws X cards."
//
// S20 sub-PR 3: a targeted X spell. Drawing more cards than the
// library holds flags the drawer for the CR 704.5b loss at the next
// SBA check — the classic Stroke kill.
func init() {
	Register(Spec{
		OracleID:     "0cc6d683-366f-4ae4-be60-20ad9621fdaf",
		Name:         "Stroke of Genius",
		Completeness: CompletenessFull,
		XMatters:     true,
		Targets:      TargetPlayer("target player"),
		OnResolve:    targetPlayerDrawsX(),
	})
}

// targetPlayerDrawsX is "Target player draws X cards" (Stroke of Genius,
// Braingeyser): the target, if still legal, draws X. A constructor, so
// the XMatters guard sees each card call it.
func targetPlayerDrawsX() func(*game.StackItem, *Context) error {
	return func(_ *game.StackItem, ctx *Context) error {
		if p, ok := firstLegalPlayerTarget(ctx); ok {
			return DrawCards{Player: p, N: ctx.X()}.Apply(ctx)
		}
		return nil
	}
}
