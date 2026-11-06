package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pearl of Wisdom — Sorcery {2}{U} (Bloomburrow, #4467):
//
//	"This spell costs {1} less to cast if you control an Otter.
//	 Draw two cards."
//
// The reduction is a self cost modifier (Winged Words' and Saved by the
// Shell's shape), read when the total cost is determined (CR 601.2f) and
// taken off the generic part only (CR 118.7a), so with an Otter the spell
// costs {1}{U}.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "91ddd2bd-2dec-4670-8edb-e89c009b2187",
		Name:         "Pearl of Wisdom",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 2},
		SelfCostModifiers: []game.CostModifier{
			CostsLess(1, "This spell costs {1} less to cast if you control an Otter.", YouControlA(HasSubtype("Otter"))),
		},
		OnResolve: ResolveSteps(DrawCards{N: 2}),
	})
}
