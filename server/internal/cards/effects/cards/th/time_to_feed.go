package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Time to Feed — #4380, S76. Uses the existing catalog seams.
func init() {
	effects.Register(effects.Spec{
		OracleID: "59642d6f-8a25-4f79-9e81-643eac775658", Name: "Time to Feed", Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()), effects.TargetCreature("target creature you control", effects.YouControl())),
		OnResolve: effects.ResolveSteps(effects.Instruction(func(ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok {
				return nil
			}
			return effects.WhenThatCreatureDiesThisTurn(ctx, t.ID, "Time to Feed — gain 3 life", timeToFeedLife)
		}), effects.Instruction(effects.FightTargetClauses)),
	})
}

var timeToFeedLife = game.SimpleDelayedBody("time-to-feed/gain-three-life", func(g *game.Game, item *game.StackItem) error {
	return effects.GainLife{Amount: 3}.Apply(effects.NewContext(g, item))
})
