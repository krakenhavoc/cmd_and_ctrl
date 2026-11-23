package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// No Way Out — Sorcery {2}{B} (Midnight Hunt):
//
//	"Target opponent discards two cards. You create a 2/2 black Zombie
//	 creature token with decayed."
//
// The token waits for the discard choice (ResolveSteps, CR 608.2c) and
// is made even if the opponent holds fewer than two cards.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8cbbb1f4-4654-4865-a14c-0c94cd5d8b19",
		Name:         "No Way Out",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.WithRiders(effects.ForTargets(game.TargetPurpose{Slot: 0, Discards: 2}), game.Purpose{Tokens: 1, CreatureTokens: 1}),
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				t, ok := ctx.ClauseTarget(0)
				if !ok || t.Kind != game.TargetPlayer {
					return nil
				}
				ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: t.ID, Source: ctx.Item.SourceCardID, N: 2})
				return nil
			}),
			effects.CreateToken{Template: effects.TokenCard("2/2 black Zombie with decayed"), N: 1},
		),
	})
}
