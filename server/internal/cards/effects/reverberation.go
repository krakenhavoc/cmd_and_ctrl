package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Reverberation — Instant {2}{U}{U}:
//
//	"All damage that would be dealt this turn by target sorcery spell is
//	 dealt to that spell's controller instead."
//
// ADR 0108 §9 (#1905): a redirection for the rest of the turn of the
// targeted spell's damage, to anything, dealt to its controller. The
// ruling: a sorcery that deals damage to several things deals its
// controller the sum.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "d3088e1d-62c9-4478-9ef7-fc3c9e5cfadb",
		Name:         "Reverberation",
		Completeness: CompletenessFull,
		Targets:      TargetSpell("target sorcery spell", Sorcery()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok {
				return nil
			}
			return RedirectDamage{From: t.ID, Protect: ShieldAnything, To: RedirectToSourceController}.Apply(ctx)
		},
	})
}
