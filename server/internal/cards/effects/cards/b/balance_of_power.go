package cardsb

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Balance of Power — Sorcery {3}{U}{U} (Eighth Edition, S82 older sets batch 618, #5607):
//
//	"If target opponent has more cards in hand than you, draw cards equal to
//	 the difference."
//
// The two hands are counted as this resolves, so the spell itself (already off
// the stack) is not in the caster's hand. A target that has become illegal
// does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c57ee68e-6832-4d4c-b710-fc03a2c10a9c",
		Name:         "Balance of Power",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			victim := effects.TargetedPlayerOf(ctx)
			if victim == uuid.Nil {
				return nil
			}
			diff := effects.CardsInHand(ctx.Game, victim) - effects.CardsInHand(ctx.Game, ctx.Controller())
			if diff <= 0 {
				return nil
			}
			return effects.DrawCards{Player: ctx.Controller(), N: diff}.Apply(ctx)
		},
	})
}
