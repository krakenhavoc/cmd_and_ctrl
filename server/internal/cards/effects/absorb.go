package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Absorb — Instant {W}{U}{U} (EDHREC rank 3716):
//
//	"Counter target spell. You gain 3 life."
//
// Cancel with a rider, the Azorius mirror of Countersquall. The
// announced spell is countered if it is still on the stack, and the
// caster gains 3 either way — the life is not conditional on the
// counter taking, so a spell that can't be countered still pays it
// out, as printed. With the target gone before resolution the whole
// spell fizzles (CR 608.2b) and no life is gained, which is also as
// printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "132ca99a-a3c7-4ed6-b4d0-0edcd7140ca2",
		Name:         "Absorb",
		Completeness: CompletenessFull,
		Targets:      TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return b35CounterChosenSpellThenGainLife(item, ctx, 3)
		},
	})
}
