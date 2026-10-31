package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bionic Blow — {X}{R}{R} Sorcery (Marvel Super Heroes Commander,
// #4455):
//
//	"Target creature you control gets +X/+0 until end of turn. Then it
//	 deals damage equal to its power to up to one other target
//	 creature."
//
// Wolf Strike's pump-then-bite: two target clauses read positionally
// (slot 0 pumps and deals, slot 1 takes), the second "up to one" and
// Distinct so it cannot be the first. The layers are recomputed after
// the pump so the bite reads the pumped power, and oneSidedBite is the
// shared one-sided damage body. Slot 1 is empty when no victim was
// chosen, and then only the pump happens.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "27683e43-7c3f-45b3-8ca4-968428000315",
		Name:         "Bionic Blow",
		Completeness: CompletenessFull,
		XMatters:     true,
		Purpose:      ForTargets(BenefitUntilEOT(0, 0, 0)),
		Targets: Clauses(
			TargetCreature("target creature you control", YouControl()),
			Distinct(TargetCreature("up to one other target creature").WithCount(0, 1)),
		),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				t, ok := ctx.ClauseTarget(0)
				if !ok || t.Kind != game.TargetCard {
					return nil
				}
				return BoostUntilEOT{Target: t.ID, Power: ctx.X(), Label: "Bionic Blow — +X/+0"}.Apply(ctx)
			}),
			Instruction(func(ctx *Context) error {
				// The bite reads the biter's power as it now stands.
				ctx.Game.RecomputeLayersIfStaleLocked()
				return oneSidedBite(ctx.Item, ctx)
			}),
		),
	})
}
