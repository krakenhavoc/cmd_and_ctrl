package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Immoral Bargain — Sorcery {1}{B}{G}:
//
//	"As an additional cost to cast this spell, sacrifice X creatures.
//	 Destroy X target nonland permanents."
//
// Eliminate the Competition's shape over a wider target clause: the
// announced X is the sacrifice count and the target count at once
// (CR 107.3a, 107.3i, ADR 0100 §3), and the destruction is one
// simultaneous event over the targets still legal (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "c82cc78b-ffd6-4f72-881d-85913830feb2",
		Name:           "Immoral Bargain",
		Completeness:   CompletenessFull,
		AdditionalCost: SacrificeXCost("X creatures", Creature()),
		Targets:        vsXTarget("X target nonland permanents", Nonland()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return vsDestroyTheTargets(ctx)
		},
	})
}
