package cardsv

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Viconia, Drow Apostate — {2}{B} Legendary Creature — Elf Cleric 2/3
//
//	"At the beginning of your upkeep, if there are four or more creature
//	 cards in your graveyard, return a creature card at random from your
//	 graveyard to your hand.
//	 Choose a Background"
//
// An intervening "if" (CR 603.4): read as the upkeep begins and again as the
// trigger resolves, so a graveyard that has dropped below four does nothing.
// The card is picked uniformly at random among the creature cards there as
// it resolves. "Choose a Background" is the deck-construction rule
// (CR 702.124k) read from the oracle text.
func init() {
	atLeastFour := func(g *game.Game, you uuid.UUID) bool {
		return len(effects.GraveyardIDs(g, you, func(c game.Card) bool { return c.IsCreature() })) >= 4
	}
	trigger := effects.AtYourUpkeep("Viconia, Drow Apostate — return a creature card at random from your graveyard to your hand",
		func(g *game.Game, item *game.StackItem) error {
			ctx := effects.NewContext(g, item)
			ids := effects.GraveyardIDs(g, item.Controller, func(c game.Card) bool { return c.IsCreature() })
			picked := effects.RandomPick(ctx, ids, 1)
			if len(picked) == 0 {
				return nil
			}
			return g.ReturnFromGraveyardForEffect(picked[0], game.ZoneHand)
		})
	effects.Register(effects.Spec{
		OracleID:     "22e36922-a81a-4a0c-b698-51b48bd91ce6",
		Name:         "Viconia, Drow Apostate",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.IfVentureCond(atLeastFour, trigger)},
	})
}
