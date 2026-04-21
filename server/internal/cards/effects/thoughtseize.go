package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thoughtseize — "Target player reveals their hand. You choose a
// nonland, nonThoughtseize card from it. That player discards that
// card. You lose 2 life."
//
// S14 sandbox simplifications:
//   - Reveal: every card in the target player's hand becomes known
//     to all seated players via AddKnower. Sticky — viewers keep
//     the knowledge after the reveal.
//   - Choice: no "you pick" UI in S14. One random hand card is
//     discarded. A richer picker (click a revealed card) lands in
//     S22 library-manipulation polish.
//   - Cost: the 2 life loss fires unconditionally.
func init() {
	Register(Spec{
		OracleID:   "edd8d1e8-be43-4c38-bb3a-83081fbaf0b5",
		Name:       "Thoughtseize",
		TargetMode: "player",
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
				return nil
			}
			target := item.Targets[0].ID
			ctx.Game.RevealHandForEffect(target)
			if err := (DiscardCards{Player: target, N: 1}).Apply(ctx); err != nil {
				return err
			}
			return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), ctx.Controller(), -2)
		},
	})
}
