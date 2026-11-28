package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Insidious Bookworms — Creature — Worm {B} 1/1 (Alliances, S82 OLD batch
// 819):
//
//	"When this creature dies, you may pay {1}{B}. If you do, target player
//	 discards a card at random."
//
// The player is chosen as the ability is put on the stack; the payment is
// made on resolution, and the discard is theirs to suffer only if it is.
func init() {
	const label = "Insidious Bookworms — you may pay {1}{B}; if you do, target player discards a card at random"
	t := effects.WhenThisDies(label, func(g *game.Game, item *game.StackItem) error {
		return effects.MayPay{Chooser: item.Controller, Cost: "{1}{B}", Question: label, Then: insidiousBookwormsDiscard}.Apply(effects.NewContext(g, item))
	})
	t.Targets = effects.TargetPlayer("target player")
	effects.Register(effects.Spec{
		OracleID:     "ff3eeed9-6801-4cab-a97a-72f5f28acb24",
		Name:         "Insidious Bookworms",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{t},
	})
}

var insidiousBookwormsDiscard = effects.PayUnlessBranches("pay-unless/insidious-bookworms-discard", func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
	player := effects.TargetedPlayerOf(ctx)
	if player == uuid.Nil {
		return nil
	}
	return ctx.Game.DiscardRandomKeyedForEffect(player, 1, game.KeyedTail{})
}, nil)
