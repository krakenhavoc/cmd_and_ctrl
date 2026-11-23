package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Peer Past the Veil — Instant {2}{R}{G} (Duskmourn: House of Horror,
// #4466):
//
//	"Discard your hand. Then draw X cards, where X is the number of card
//	 types among cards in your graveyard."
//
// Change of Fortune's two instructions with a different count. The whole
// hand goes first, so the cards just discarded count toward X; X is read
// after the discard, off the graveyard as it then stands
// (b16CardTypesInGraveyard, the delirium count). Peer Past the Veil is
// still on the stack while it resolves, so it does not count itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8fb7ad08-9bb2-49f1-8ed0-8cc8abca4fe0",
		Name:         "Peer Past the Veil",
		Completeness: effects.CompletenessFull,
		// The draw waits for a prompt a discard asks (a replacement offer
		// such as Library of Leng's; CR 608.2c, #4155).
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				_, err := effects.DiscardWholeHand(ctx.Game, ctx.Controller())
				return err
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				caster := ctx.Controller()
				return effects.DrawCards{Player: caster, N: effects.B16CardTypesInGraveyard(ctx.Game, caster)}.Apply(ctx)
			}),
		),
	})
}
