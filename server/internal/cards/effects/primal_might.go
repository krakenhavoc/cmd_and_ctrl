package effects

// Primal Might — #4380, S76. Uses the existing catalog seams.
func init() {
	Register(Spec{
		OracleID: "58d7b8b3-ac47-4805-8259-4ba8e70d7dc3", Name: "Primal Might", Completeness: CompletenessFull,
		XMatters: true, Purpose: ForTargets(BenefitUntilEOT(0, 0, 0)),
		Targets: Clauses(TargetCreature("target creature you control", YouControl()), TargetCreature("up to one target creature you don't control", Not(YouControl())).WithCount(0, 1)),
		OnResolve: ResolveSteps(Instruction(func(ctx *Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok {
				return nil
			}
			return BoostUntilEOT{Target: t.ID, Power: ctx.X(), Toughness: ctx.X(), Label: "Primal Might — +X/+X"}.Apply(ctx)
		}), Instruction(fightTargetClauses)),
	})
}
