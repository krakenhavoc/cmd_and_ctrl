package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "d3088e1d-62c9-4478-9ef7-fc3c9e5cfadb",
		Name:         "Reverberation",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target sorcery spell", effects.Sorcery()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok {
				return nil
			}
			return effects.RedirectDamage{From: t.ID, Protect: effects.ShieldAnything, To: effects.RedirectToSourceController}.Apply(ctx)
		},
	})
}
