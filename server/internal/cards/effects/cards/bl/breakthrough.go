package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Breakthrough — Sorcery {X}{U} (Double Masters 2022, #5948):
//
//	"Draw four cards, then choose X cards in your hand and discard the rest."
//
// You keep X cards of your whole hand (not only the four drawn) and discard
// every other one, so it is "discard all but X": the discard is
// max(0, hand size - X) cards, of your choice. A hand already at or under X
// discards nothing. X is paid at cast (CR 601.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c6b26c1f-121f-4b7c-921e-429b623ca934",
		Name:         "Breakthrough",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Purpose:      game.Purpose{Draws: 4},
		OnResolve: effects.ResolveSteps(
			effects.DrawCards{N: 4},
			effects.Instruction(func(ctx *effects.Context) error {
				p := ctx.Game.PlayerByIDForEffect(ctx.Controller())
				if p == nil {
					return nil
				}
				if n := p.Hand.Size() - ctx.X(); n > 0 {
					ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: p.ID, Source: ctx.Source(), N: n})
				}
				return nil
			}),
		),
	})
}
