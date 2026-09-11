package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blasphemous Act — Sorcery {8}{R}:
//
//	"This spell costs {1} less to cast for each creature on the
//	 battlefield. Blasphemous Act deals 13 damage to each creature."
//
// S23 fixes a real bug as well as reworking the sweep: the S14 entry
// DESTROYED every creature, and this card does not destroy anything.
// It deals damage, which indestructible survives, damage prevention
// stops, and lifelink profits from — and which kills through the
// lethal-damage state-based action rather than immediately, so the
// deaths land on the SBA sweep's simultaneous batch instead of
// mid-resolution. Thirteen damage kills the same board in practice
// and the difference was invisible in a test with no indestructible
// creature, which is exactly how this kind of bug survives.
//
// Sandbox simplification still standing: THE COST REDUCTION IS NOT
// IMPLEMENTED — this casts at its printed {8}{R}. Cost modification
// has no Spec hook (S28 territory), and lowering the printed cost
// would be wrong in the other direction.
func init() {
	Register(Spec{
		OracleID:     "7a2484a9-04fd-41a0-8224-610c1c07ed10",
		Name:         "Blasphemous Act",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The cost reduction is missing — it always costs the full {8}{R} no matter how many creatures are on the battlefield."},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return damageEachMatching(ctx, Creature(), 13)
		},
	})
}
