package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crypt Lurker — Creature — Horror {3}{B}, 3/4 (Core Set 2021, #5511):
//
//	"When this creature enters, you may sacrifice a creature or discard a
//	 creature card. If you do, draw a card."
//
// The either/or may-pay (K'un-Lun Warrior's shape): one prompt whose answer
// names the branch, a branch the player can't make is never offered, and a
// player with neither a creature nor a creature card in hand is not asked.
// The Lurker itself is a creature it may sacrifice. The discard branch takes
// only a creature card (PayAction.CardType).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f0cbf361-d7dc-4556-a654-71c08785095a",
		Name:         "Crypt Lurker",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Crypt Lurker — you may sacrifice a creature or discard a creature card; if you do, draw a card",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Pay:      effects.S82OldB105SacrificeCreatureOrDiscardCreatureCard(),
						Question: "Crypt Lurker — sacrifice a creature or discard a creature card to draw a card?",
						Then:     effects.MayPayDrawACard,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
