package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bazaar Trademage — Creature — Human Wizard {2}{U}, 3/4 (Modern Horizons):
//
//	"Flying
//	 When this creature enters, draw two cards, then discard three cards."
//
// Flying is a printed keyword. The enters trigger draws two and then makes its
// controller discard three of the cards they hold (CR 701.8), the whole hand
// if it has fewer; the discard waits for the draw, so a card drawn can be one
// of those discarded.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6c94a255-c031-48d7-aa07-1189ad3da3b6",
		Name:            "Bazaar Trademage",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{Draws: 2},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Bazaar Trademage — draw two cards, then discard three cards",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if err := (effects.DrawCards{N: 2}).Apply(ctx); err != nil {
						return err
					}
					g.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: ctx.Controller(), Source: ctx.Source(), N: 3})
					return nil
				}),
		},
	})
}
