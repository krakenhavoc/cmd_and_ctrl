package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Curious Herd — Instant {3}{G} (Commander 2020, #5513):
//
//	"Choose target opponent. You create X 3/3 green Beast creature tokens,
//	 where X is the number of artifacts that player controls."
//
// The opponent is the spell's only target; X is counted as the spell
// resolves (CR 608.2h), every artifact that player controls counting, an
// artifact creature or a token included. An opponent who has left by then
// is an illegal target and the spell does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e6f2e363-e361-4571-af37-6c63da04bdbf",
		Name:         "Curious Herd",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			opp, ok := effects.FirstLegalPlayerTarget(ctx)
			if !ok {
				return nil
			}
			n := 0
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.Controller == opp && c.IsArtifact() {
					n++
				}
			}
			if n == 0 {
				return nil
			}
			return effects.CreateToken{Template: effects.TokenCard("3/3 green Beast"), N: n}.Apply(ctx)
		},
	})
}
