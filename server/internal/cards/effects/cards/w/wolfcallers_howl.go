package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wolfcaller's Howl — Enchantment {3}{G} (Commander 2014, #5553):
//
//	"At the beginning of your upkeep, create X 2/2 green Wolf creature
//	 tokens, where X is the number of your opponents with four or more cards
//	 in hand."
//
// It always triggers; X is counted as the ability resolves, over the
// opponents still in the game, so an opponent who discards down in response
// is not counted. With no such opponent it makes nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e4383685-04e2-4516-b089-863c8d309250",
		Name:         "Wolfcaller's Howl",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Wolfcaller's Howl — create a 2/2 green Wolf for each opponent with four or more cards in hand",
				effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					n := wolfcallersHowlOpponentsWithFourCards(ctx)
					if n <= 0 {
						return nil
					}
					return effects.CreateToken{Template: effects.TokenCard("2/2 green Wolf"), N: n}.Apply(ctx)
				}))),
		},
	})
}

// wolfcallersHowlOpponentsWithFourCards counts the controller's opponents
// with four or more cards in hand.
func wolfcallersHowlOpponentsWithFourCards(ctx *effects.Context) int {
	n := 0
	for _, id := range ctx.Opponents() {
		if p := ctx.Game.PlayerByIDForEffect(id); p != nil && p.Hand.Size() >= 4 {
			n++
		}
	}
	return n
}
