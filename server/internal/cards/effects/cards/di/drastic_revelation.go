package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Drastic Revelation — Sorcery {2}{U}{B}{R} (Alara Reborn, #5586):
//
//	"Discard your hand. Draw seven cards, then discard three cards at
//	 random."
//
// Three steps in order (CR 608.2c): the whole hand is discarded, seven cards
// are drawn (each its own draw, CR 121.2), and then three cards are discarded
// at random from the new hand; a smaller hand is emptied.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cc519689-031c-48ba-9c79-3f5a435a0848",
		Name:         "Drastic Revelation",
		Completeness: effects.CompletenessFull,
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				_, err := effects.DiscardWholeHand(ctx.Game, ctx.Controller())
				return err
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.DrawCards{N: 7}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.DiscardCards{N: 3}.Apply(ctx)
			}),
		),
	})
}
