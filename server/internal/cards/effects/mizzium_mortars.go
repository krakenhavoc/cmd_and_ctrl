package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mizzium Mortars — Sorcery {1}{R} (EDHREC rank 3003):
//
//	"Mizzium Mortars deals 4 damage to target creature you don't
//	 control.
//	 Overload {3}{R}{R}{R} (You may cast this spell for its overload
//	 cost. If you do, change "target" in its text to "each.")"
//
// Red's one-sided sweeper. Vandalblast's shape: the overload
// machinery carries both halves — the {3}{R}{R}{R} paid instead of
// {1}{R}, and the deletion of the target clause, which turns "target
// creature you don't control" into "each creature you don't
// control". The predicate — And(Creature(), OpponentControls()) — is
// the target clause in one mode and the sweep's filter in the other,
// applied at resolution to whatever is on the battlefield then, and
// the damage is the spell's in both. Your own creatures are never a
// legal target and never in the sweep.
//
// No simplification.
func init() {
	creatureYouDontControl := And(Creature(), OpponentControls())
	Register(Spec{
		OracleID:     "48ddba1e-2ad7-463f-9307-d2379a800e51",
		Name:         "Mizzium Mortars",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature you don't control", OpponentControls()),
		AlternativeCosts: []game.AlternativeCost{
			Overload("{3}{R}{R}{R}"),
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if ctx.PaidAltCost("overload") {
				return damageEachMatching(ctx, creatureYouDontControl, 4)
			}
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return DealDamage{Source: ctx.Source(), Target: id, Amount: 4}.Apply(ctx)
		},
	})
}
