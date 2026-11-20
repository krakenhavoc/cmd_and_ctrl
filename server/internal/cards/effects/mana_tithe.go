package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mana Tithe — Instant {W} (Planar Chaos, #5595):
//
//	"Counter target spell unless its controller pays {1}."
//
// Force Spike in white: the pay-unless prompt goes to the targeted spell's
// controller (CR 118.12), and a payment they can't fund counters it. A
// target gone by resolution does nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "948217ac-289e-455f-8c7c-519aa75fa345",
		Name:         "Mana Tithe",
		Completeness: CompletenessFull,
		Targets:      TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return counterTheTargetSpellUnlessPaid("Mana Tithe", "{1}")(ctx.Game, item)
		},
	})
}
