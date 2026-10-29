package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Time to Feed — #4380, S76. Uses the existing catalog seams.
func init() {
	Register(Spec{
		OracleID: "59642d6f-8a25-4f79-9e81-643eac775658", Name: "Time to Feed", Completeness: CompletenessFull,
		Targets: Clauses(TargetCreature("target creature an opponent controls", OpponentControls()), TargetCreature("target creature you control", YouControl())),
		OnResolve: ResolveSteps(Instruction(func(ctx *Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok {
				return nil
			}
			return whenThatCreatureDiesThisTurn(ctx, t.ID, "Time to Feed — gain 3 life", timeToFeedLife)
		}), Instruction(fightTargetClauses)),
	})
}

var timeToFeedLife = game.SimpleDelayedBody("time-to-feed/gain-three-life", func(g *game.Game, item *game.StackItem) error { return GainLife{Amount: 3}.Apply(NewContext(g, item)) })
