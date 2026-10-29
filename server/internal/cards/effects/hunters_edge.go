package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hunter's Edge — #4380, S76. Uses the existing catalog seams.
func init() {
	Register(Spec{
		OracleID: "7b428e94-8428-455f-b555-f648fcdd2127", Name: "Hunter's Edge", Completeness: CompletenessFull,
		Targets: Clauses(TargetCreature("target creature you control", YouControl()), TargetCreature("target creature you don't control", Not(YouControl()))),
		OnResolve: ResolveSteps(Instruction(counterOnFirstCreatureTarget), Instruction(func(ctx *Context) error {
			mine, ok := ctx.ClauseTarget(0)
			if !ok {
				return nil
			}
			other, ok := ctx.ClauseTarget(1)
			if !ok {
				return nil
			}
			ref, ok := ctx.Game.PermanentRefForEffect(mine.ID)
			if !ok {
				return nil
			}
			info, ok := ctx.Game.PermanentForEffect(ref)
			if !ok {
				return nil
			}
			return DealDamage{SourceObject: &ref, Target: other.ID, Amount: info.Power}.Apply(ctx)
		})),
	})
}

func counterOnFirstCreatureTarget(ctx *Context) error {
	t, ok := ctx.ClauseTarget(0)
	if !ok {
		return nil
	}
	return AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
}
