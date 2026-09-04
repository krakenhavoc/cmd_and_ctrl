package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mind Rot — "Target player discards two cards."
//
// Per CR 701.8a, the player who's discarding chooses the cards
// unless the effect says "at random." Mind Rot's text doesn't;
// the target chooses. Queue a pending discard-choice into the
// shared DiscardPending map; the target's client shows the
// S13.4 DiscardPromptModal and submits their picks via
// discard_selection. The spell still resolves to graveyard right
// now — only the hand-picking is deferred.
func init() {
	Register(Spec{
		OracleID: "ad44cf74-b717-48fb-9fa2-77512024d76a",
		Name:     "Mind Rot",
		Targets:  TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
				return nil
			}
			ctx.Game.DiscardChoiceForEffect(item.Targets[0].ID, 2)
			return nil
		},
	})
}
