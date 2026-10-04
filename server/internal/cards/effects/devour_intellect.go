package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Devour Intellect — Sorcery {B}:
//
//	"Target opponent discards a card. If mana from a Treasure was spent
//	 to cast this spell, instead that player reveals their hand, you
//	 choose a nonland card from it, then that player discards that
//	 card."
//
// The Treasure check reads the spell's own payment record
// (Context.ManaSpent().FromTreasure(), #1212) as it resolves, the same
// fact Hired Hexblade reads. Without Treasure mana the opponent chooses
// their own discard (CR 701.9b). With it the revealed-hand pick
// (ADR 0116) replaces that, with Thoughtseize's nonland filter. A
// payment the engine did not record spends no Treasure mana it knows
// of, so the plain discard is the fallback: weaker, never stronger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "abf8264e-d020-4939-8092-83469b1a2244",
		Name:         "Devour Intellect",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target opponent", Opponent()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			player := TargetedPlayer(ctx)
			if player == uuid.Nil {
				return nil
			}
			if ctx.ManaSpent().FromTreasure() {
				return ChooseFromRevealedHand{Player: player, Filter: Nonland(), Label: "nonland card"}.Apply(ctx)
			}
			ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{
				Player: player,
				Source: item.SourceCardID,
				N:      1,
			})
			return nil
		},
	})
}
