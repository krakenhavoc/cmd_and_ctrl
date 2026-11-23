package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kassandra, Eagle Bearer — Legendary Creature — Human Assassin Warrior
// {1}{R}{W}, 2/2 (Assassin's Creed, #5948):
//
//	"Haste
//	 When Kassandra enters, search your graveyard, hand, and library for a
//	 card named The Spear of Leonidas, put it onto the battlefield, then
//	 shuffle.
//	 Whenever a creature you control with a legendary Equipment attached to
//	 it deals combat damage to a player, draw a card."
//
// The search is Shaun & Rebecca's three-zone search by name (AlsoGraveyardAndHand:
// all three zones are named, so the library is always searched and shuffled; a
// hidden zone may fail to find, CR 701.23b). The trigger is one per creature
// that connects, Ezio, Blade of Vengeance's read of the damage event, with the
// Equipment read off the dealer as it dealt the damage: a legendary Equipment
// attached to it at that moment (CR 301.5), whoever controls the Equipment.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "745b9a2c-bcd7-4db9-8cd0-d38be2ef6270",
		Name:            "Kassandra, Eagle Bearer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Kassandra, Eagle Bearer — search your graveyard, hand, and library for a card named The Spear of Leonidas and put it onto the battlefield, then shuffle",
				func(g *game.Game, item *game.StackItem) error {
					return effects.SearchLibrary{
						Player:               item.Controller,
						Predicate:            func(c game.Card) bool { return c.Name == "The Spear of Leonidas" },
						Dest:                 game.ZoneBattlefield,
						Limit:                1,
						AlsoGraveyardAndHand: true,
						Shuffle:              true,
						Reason:               "Kassandra, Eagle Bearer — search your graveyard, hand, and library for a card named The Spear of Leonidas",
					}.Apply(effects.NewContext(g, item))
				}),
			effects.On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if !effects.CombatDamageToPlayerBy(ev, source.Controller, g) {
					return false
				}
				equipment, legend := effects.HasSubtype("Equipment"), effects.Legendary()
				for _, c := range g.Battlefield.Cards {
					if c.AttachedTo.ID == ev.Source && equipment(g, source.Controller, c) && legend(g, source.Controller, c) {
						return true
					}
				}
				return false
			}, "Kassandra, Eagle Bearer — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
