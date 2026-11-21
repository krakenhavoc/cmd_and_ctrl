package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bionic Blow — {X}{R}{R} Sorcery (Marvel Super Heroes Commander,
// #4455):
//
//	"Target creature you control gets +X/+0 until end of turn. Then it
//	 deals damage equal to its power to up to one other target
//	 creature."
//
// Wolf Strike's pump-then-bite: two target clauses read positionally
// (slot 0 pumps and deals, slot 1 takes), the second "up to one" and
// Distinct so it cannot be the first. The bite reads the pumped power,
// since a read during the resolution sees the pump before it (#4909),
// and oneSidedBite is the shared one-sided damage body. Slot 1 is empty
// when no victim was chosen, and then only the pump happens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "27683e43-7c3f-45b3-8ca4-968428000315",
		Name:         "Bionic Blow",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Purpose:      effects.ForTargets(effects.PowerXUntilEOT(0)),
		Targets: effects.Clauses(
			effects.TargetCreature("target creature you control", effects.YouControl()),
			effects.Distinct(effects.TargetCreature("up to one other target creature").WithCount(0, 1)),
		),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				t, ok := ctx.ClauseTarget(0)
				if !ok || t.Kind != game.TargetCard {
					return nil
				}
				return effects.BoostUntilEOT{Target: t.ID, Power: ctx.X(), Label: "Bionic Blow — +X/+0"}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.OneSidedBite(ctx.Item, ctx)
			}),
		),
	})
}
