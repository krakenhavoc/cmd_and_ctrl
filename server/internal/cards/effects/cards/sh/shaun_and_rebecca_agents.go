package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shaun & Rebecca, Agents — Legendary Creature — Human Assassin Scientist
// {1}{G}{W}{U}, 4/4 (Assassin's Creed, #5948):
//
//	"Vigilance
//	 When Shaun & Rebecca enters, search your graveyard, hand, and library
//	 for a card named The Animus and put it onto the battlefield, then
//	 shuffle.
//	 {T}: Add {C}. When you do, mill two cards."
//
// The search is Kotose's three-zone search (AlsoGraveyardAndHand: all three
// zones are named, so the library is always searched and shuffled; a hidden
// zone may fail to find, CR 701.23b), taking one card by name. The mana ability
// is {T}: Add {C}; "When you do, mill two cards" is Rubble Rouser's reflexive
// shape: an ordinary triggered ability (CR 605.1b, damage-and-mill can't add
// mana), watching this row's activation, so tapping her to attack, crew or by
// an opponent's effect mills nothing. The {C} lands at once and the mill goes
// on the stack the next time a player would receive priority.
//
// No simplification.
func init() {
	const (
		manaLabel = "{T}: Add {C}. When you do, mill two cards"
		millLabel = "Shaun & Rebecca, Agents — mill two cards"
	)
	effects.Register(effects.Spec{
		OracleID:        "795a2b09-2ed1-4bc8-b6c9-fb26a81e6790",
		Name:            "Shaun & Rebecca, Agents",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    manaLabel,
		}},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Shaun & Rebecca, Agents — search your graveyard, hand, and library for a card named The Animus and put it onto the battlefield, then shuffle",
				func(g *game.Game, item *game.StackItem) error {
					return effects.SearchLibrary{
						Player:               item.Controller,
						Predicate:            func(c game.Card) bool { return c.Name == "The Animus" },
						Dest:                 game.ZoneBattlefield,
						Limit:                1,
						AlsoGraveyardAndHand: true,
						Shuffle:              true,
						Reason:               "Shaun & Rebecca, Agents — search your graveyard, hand, and library for a card named The Animus",
					}.Apply(effects.NewContext(g, item))
				}),
			effects.On(game.EventManaAbilityActivated,
				func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return source != nil && ev.Source == source.InstanceID && ev.Label == manaLabel && ev.Actor == source.Controller
				},
				millLabel, effects.Do(effects.MillCards{N: 2})),
		},
	})
}
