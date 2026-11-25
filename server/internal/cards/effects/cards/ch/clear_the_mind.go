package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Clear the Mind — Sorcery {2}{U} (Ravnica Allegiance):
//
//	"Target player shuffles their graveyard into their library.
//	 Draw a card."
//
// The target's graveyard goes into the target's library and is shuffled; the
// spell's controller then draws, whoever the target was. If the target is no
// longer legal the whole spell does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "141c33be-7a25-4981-b62c-efee8e8ec91d",
		Name:         "Clear the Mind",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		Purpose:      game.Purpose{Draws: 1},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer || !ctx.IsTargetLegal(item.Targets[0]) {
				return nil
			}
			if err := effects.FinaleShuffleGraveyardIntoLibrary(ctx, item.Targets[0].ID); err != nil {
				return err
			}
			return effects.DrawCards{N: 1}.Apply(ctx)
		},
	})
}
