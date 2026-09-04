package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thoughtseize — "Target player reveals their hand. You choose a
// nonland, nonThoughtseize card from it. That player discards that
// card. You lose 2 life."
//
// Flow:
//  1. Reveal the target's hand to the caster via
//     QueueDiscardFromRevealedHand. Sticky per S13.5 KnownBy.
//  2. Queue a PendingChoice with Kind=discard_from_hand,
//     chooser=caster, fromPlayer=target, count=1. The caster's
//     client pops a picker modal showing the target's hand; the
//     caster submits their pick via resolve_choice. The target's
//     card moves to their graveyard.
//  3. The caster loses 2 life immediately — the cost is part of
//     resolution, not the deferred choice.
//
// The "nonland, nonThoughtseize" predicate is not enforced by the
// engine; the caster is expected to pick something legal. S20
// smart-cast UI can validate at submit time.
func init() {
	Register(Spec{
		OracleID: "edd8d1e8-be43-4c38-bb3a-83081fbaf0b5",
		Name:     "Thoughtseize",
		Targets:  TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
				return nil
			}
			target := item.Targets[0].ID
			ctx.Game.QueueDiscardFromRevealedHand(
				ctx.Controller(), target, ctx.Source(),
				1, "Thoughtseize",
			)
			return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), ctx.Controller(), -2)
		},
	})
}
