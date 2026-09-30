package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Eliminate the Competition — Sorcery {4}{B}:
//
//	"As an additional cost to cast this spell, sacrifice X creatures.
//	 Destroy X target creatures."
//
// "Sacrifice X" as a cast's additional cost (ADR 0100 §3): the X is
// announced with the cast (CR 107.3a), the caster names exactly that
// many creatures to sacrifice, and CR 107.3i makes it the same X the
// target clause counts — so the announce path pins the targets to X
// too. The sacrifices happen with the spell on the stack; the
// destruction is one simultaneous event, and a target that left in
// response is skipped (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "26e549c0-a08b-475b-9138-6dde175cdf55",
		Name:           "Eliminate the Competition",
		Completeness:   CompletenessFull,
		AdditionalCost: SacrificeXCost("X creatures", Creature()),
		Targets:        vsXTarget("X target creatures", Creature()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return vsDestroyTheTargets(ctx)
		},
	})
}
