package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bridge from Below — {B}{B}{B} Enchantment (Special Guests):
//
//	"Whenever a nontoken creature is put into your graveyard from the
//	 battlefield, if this card is in your graveyard, create a 2/2 black
//	 Zombie creature token.
//	 When a creature is put into an opponent's graveyard from the
//	 battlefield, if this card is in your graveyard, exile this card."
//
// Two triggers that work only from the graveyard (InGraveyard, CR 113.6b),
// both on a creature that went from the battlefield to a graveyard. "Your
// graveyard" is the graveyard of the creature card's OWNER, not of whoever
// controlled it, and the creature is read as it last existed (CR 603.10a).
// The card must already have been in the graveyard: one that arrives in the
// same batch as the creature (a wrath that takes both) does not trigger,
// which is the printed ruling. "If this card is in your graveyard" is an
// intervening if (CR 603.4): the zone gate is the trigger-time half, and as
// the ability resolves the card must still be there as the same object
// (CondThisInYourGraveyard), so a second trigger from one batch of deaths
// does nothing once the first has exiled it. The token is made for every
// nontoken creature, so two of yours dying together make two Zombies.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0f071f64-b69b-4fa0-999c-5028429e3cfb",
		Name:         "Bridge from Below",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondThisInYourGraveyard{}, effects.InGraveyard(effects.On(game.EventLTB, bridgeFromBelowYourNontokenCreatureDied,
				"Bridge from Below — create a 2/2 black Zombie creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 black Zombie"), N: 1})))),
			effects.InterveningIf(effects.CondThisInYourGraveyard{}, effects.InGraveyard(effects.On(game.EventLTB, bridgeFromBelowOpponentsCreatureDied,
				"Bridge from Below — exile this card",
				func(g *game.Game, item *game.StackItem) error {
					if effects.SourceIsNewObject(g, item) {
						return nil
					}
					return g.ExileCardForEffect(item.SourceCardID)
				}))),
		},
	})
}

// bridgeFromBelowYourNontokenCreatureDied is "a nontoken creature is put
// into your graveyard from the battlefield" for the card in your graveyard.
func bridgeFromBelowYourNontokenCreatureDied(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
	dead, ok := effects.DiedCreature(ev, g)
	return ok && !effects.IsToken(dead) && dead.Owner == source.Controller &&
		effects.FuriousForebearWasInGraveyard(ev, source, ch, g)
}

// bridgeFromBelowOpponentsCreatureDied is "a creature is put into an
// opponent's graveyard from the battlefield".
func bridgeFromBelowOpponentsCreatureDied(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
	dead, ok := effects.DiedCreature(ev, g)
	return ok && dead.Owner != source.Controller && effects.FuriousForebearWasInGraveyard(ev, source, ch, g)
}
