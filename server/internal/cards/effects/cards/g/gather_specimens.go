package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gather Specimens — Instant {3}{U}{U}{U} (Shards of Alara, S82 OLD batch 517,
// #5588; seam #6260):
//
//	"If a creature would enter the battlefield under an opponent's control
//	 this turn, it enters under your control instead."
//
// A CR 614.12 replacement created by the resolving spell and held for the
// turn (CR 611.2): any creature that would enter under another player's
// control, by a spell, a reanimation, a flicker or a token, enters under the
// Gather Specimens controller's control instead.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e3b8d70b-17aa-4a31-a4f9-319907d088c2",
		Name:         "Gather Specimens",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			ctx.Game.CreaturesEnterUnderYourControlThisTurnForEffect(item.SourceCardID, item.Controller,
				"Gather Specimens — creatures that would enter under an opponent's control enter under your control instead")
			return nil
		},
	})
}
