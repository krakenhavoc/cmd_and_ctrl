package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scarscale Ritual — Sorcery {1}{U/B} (Shadowmoor, S82 OLD b532, #5591):
//
//	"As an additional cost to cast this spell, put a -1/-1 counter on a creature
//	 you control.
//	 Draw two cards."
//
// Lethal Sting's cost: CR 701.68a's blight 1 paid by the same component, named on
// the cast as blight_ids. A cost's counter is not doubled, and with no creature
// of yours to put it on the spell can't be cast.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "367ec7c5-8c76-4af6-b941-a3a8c2ff67ff",
		Name:           "Scarscale Ritual",
		Completeness:   effects.CompletenessFull,
		AdditionalCost: effects.MinusOneCounterOnACreatureYouControlCost(),
		Purpose:        game.Purpose{Draws: 2},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DrawCards{Player: ctx.Controller(), N: 2}.Apply(ctx)
		},
	})
}
