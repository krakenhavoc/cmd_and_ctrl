package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Case of the Crimson Pulse — Enchantment — Case {2}{R} (Murders at
// Karlov Manor, #4469):
//
//	"When this Case enters, discard a card, then draw two cards.
//	 To solve — You have no cards in hand. (If unsolved, solve at the
//	 beginning of your end step.)
//	 Solved — At the beginning of your upkeep, discard your hand, then
//	 draw two cards."
//
// The Case (ADR 0071) is Case of the Shattered Pact's skeleton. The
// enters trigger is Romantic Rendezvous' rummage: the discard is the
// controller's choice and the two cards are drawn once it is made, and
// from an empty hand nothing is discarded and two are still drawn (CR
// 701.9a). The solve condition reads the hand at the end step and again
// on resolution (CR 603.4). The solved upkeep trigger is Ox of Agonas'
// wheel, gated on the Case being solved (WhenSolved).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "955504bc-050a-42a2-b2e4-ab213183b2be",
		Name:         "Case of the Crimson Pulse",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Case of the Crimson Pulse — discard a card, then draw two cards",
				func(g *game.Game, item *game.StackItem) error {
					return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
						Player: item.Controller,
						Source: item.SourceCardID,
						N:      1,
					}, effects.MscM1RendezvousDraw, nil)
				}),
			effects.ToSolve("Case of the Crimson Pulse — to solve: you have no cards in hand", effects.R35YouHaveNoCardsInHand),
			effects.WhenSolved(effects.AtYourUpkeep("Case of the Crimson Pulse — solved: discard your hand, then draw two cards", effects.Do(
				effects.Instruction(func(ctx *effects.Context) error {
					_, err := effects.DiscardWholeHand(ctx.Game, ctx.Controller())
					return err
				}),
				effects.DrawCards{N: 2},
			))),
		},
	})
}
