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
// #746 closed the last caveat: the cost reduction is a self cost
// modifier (Spec.SelfCostModifiers, ADR 0048 addendum), counting every
// creature on the battlefield whoever controls it. It spends generic
// mana only, so the card never costs less than {R}.
func init() {
	Register(Spec{
		OracleID:     "7a2484a9-04fd-41a0-8224-610c1c07ed10",
		Name:         "Blasphemous Act",
		Completeness: CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			CostsLessEach(PermanentsOnBattlefield(Creature()),
				"This spell costs {1} less to cast for each creature on the battlefield."),
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return damageEachMatching(ctx, Creature(), 13)
		},
	})
}
