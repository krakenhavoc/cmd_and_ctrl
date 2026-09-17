package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mind Rot — "Target player discards two cards."
//
// Per CR 701.8a, the player who's discarding chooses the cards
// unless the effect says "at random." Mind Rot's text doesn't; the
// target chooses. The discard is a PendingChoice addressed to the
// target over their own hand (#651), so the table waits for it: the
// spell leaves the stack, but nobody advances a step or passes
// priority until the two cards are in the graveyard, because the
// discard is part of this resolving effect (CR 608.2c).
func init() {
	Register(Spec{
		OracleID:     "ad44cf74-b717-48fb-9fa2-77512024d76a",
		Name:         "Mind Rot",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
				return nil
			}
			ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{
				Player: item.Targets[0].ID,
				Source: item.SourceCardID,
				N:      2,
			})
			return nil
		},
	})
}
