package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sarkhan Unbroken — Legendary Planeswalker — Sarkhan {2}{G}{U}{R}, loyalty 4
// (Mythic Edition):
//
//	"+1: Draw a card, then add one mana of any color.
//	 −2: Create a 4/4 red Dragon creature token with flying.
//	 −8: Search your library for any number of Dragon creature cards, put them
//	 onto the battlefield, then shuffle."
//
// The mana is added as the ability resolves, after the draw, and empties with
// the step (CR 106.4). The −8 is an unbounded search (any number, CR 701.23d)
// for creature cards with the Dragon subtype, read through the layered view.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1d37b453-f10c-42f2-8390-626a8a8eb27c",
		Name:            "Sarkhan Unbroken",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 4,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+1: Draw a card, then add one mana of any color.",
				Cost:    effects.LoyaltyCost(1),
				Purpose: game.Purpose{Draws: 1},
				Effect:  effects.Do(effects.DrawCards{N: 1}, effects.AddMana{Produced: "{W|U|B|R|G}"}),
			},
			{
				Label:  "−2: Create a 4/4 red Dragon creature token with flying.",
				Cost:   effects.LoyaltyCost(-2),
				Effect: effects.Do(effects.CreateToken{Template: effects.TokenCard("4/4 red Dragon with flying"), N: 1}),
			},
			{
				Label: "−8: Search your library for any number of Dragon creature cards, put them onto the battlefield, then shuffle.",
				Cost:  effects.LoyaltyCost(-8),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					dragon := ctx.TextSubtype("Dragon")
					return effects.SearchLibrary{
						Player:    item.Controller,
						Predicate: func(c game.Card) bool { return c.IsCreature() && c.HasSubtypeWord(dragon) },
						Dest:      game.ZoneBattlefield,
						Unbounded: true,
						Shuffle:   true,
						Source:    item.SourceCardID,
						Reason:    "Sarkhan Unbroken — any number of Dragon creature cards onto the battlefield",
					}.Apply(ctx)
				},
			},
		},
	})
}
