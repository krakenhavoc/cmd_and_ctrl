package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mind Twist — Sorcery {X}{B} (Special Guests, S82 BNS):
//
//	"Target player discards X cards at random."
//
// Hymn to Tourach's random discard (CR 701.9b) scaled by the announced X.
// X = 0 is legal and discards nothing; a hand smaller than X is emptied.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "78f9c223-9982-4282-a496-a6f892f0a5bf",
		Name:         "Mind Twist",
		XMatters:     true,
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			player := effects.TargetedPlayer(ctx)
			if player == uuid.Nil {
				return nil
			}
			return effects.DiscardCards{Player: player, N: ctx.X()}.Apply(ctx)
		},
	})
}
