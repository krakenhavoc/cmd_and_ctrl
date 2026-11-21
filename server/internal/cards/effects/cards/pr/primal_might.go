package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Primal Might — #4380, S76. Uses the existing catalog seams.
func init() {
	effects.Register(effects.Spec{
		OracleID: "58d7b8b3-ac47-4805-8259-4ba8e70d7dc3", Name: "Primal Might", Completeness: effects.CompletenessFull,
		XMatters: true, Purpose: effects.ForTargets(effects.PumpXUntilEOT(0)),
		Targets: effects.Clauses(effects.TargetCreature("target creature you control", effects.YouControl()), effects.TargetCreature("up to one target creature you don't control", effects.Not(effects.YouControl())).WithCount(0, 1)),
		OnResolve: effects.ResolveSteps(effects.Instruction(func(ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok {
				return nil
			}
			return effects.BoostUntilEOT{Target: t.ID, Power: ctx.X(), Toughness: ctx.X(), Label: "Primal Might — +X/+X"}.Apply(ctx)
		}), effects.Instruction(effects.FightTargetClauses)),
	})
}
