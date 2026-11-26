package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nissa Revane — Legendary Planeswalker — Nissa {2}{G}{G}, loyalty 2
// (Zendikar, #5583):
//
//	"+1: Search your library for a card named Nissa's Chosen, put it onto the
//	     battlefield, then shuffle.
//	 +1: You gain 2 life for each Elf you control.
//	 −7: Search your library for any number of Elf creature cards, put them
//	     onto the battlefield, then shuffle."
//
// The searches look for a name and for a creature type, so each may find
// nothing (CR 701.23b); the life gain counts the Elves as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "113fc13a-6df9-4eed-b9d4-002a85a036a1",
		Name:            "Nissa Revane",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 2,
		Activated: []effects.ActivatedAbility{
			{
				Label: "+1: Search your library for a card named Nissa's Chosen, put it onto the battlefield, then shuffle.",
				Cost:  effects.LoyaltyCost(1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.SearchLibrary{
						Player:    item.Controller,
						Predicate: func(c game.Card) bool { return c.Name == "Nissa's Chosen" },
						Dest:      game.ZoneBattlefield,
						Limit:     1,
						Shuffle:   true,
						Source:    item.SourceCardID,
						Reason:    "Nissa Revane — a card named Nissa's Chosen",
					}.Apply(effects.NewContext(g, item))
				},
			},
			{
				Label: "+1: You gain 2 life for each Elf you control.",
				Cost:  effects.LoyaltyCost(1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.GainLife{Amount: 2 * effects.CountYouControl(ctx, effects.HasSubtype(ctx.TextSubtype("Elf")))}.Apply(ctx)
				},
			},
			{
				Label: "−7: Search your library for any number of Elf creature cards, put them onto the battlefield, then shuffle.",
				Cost:  effects.LoyaltyCost(-7),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					elf := ctx.TextSubtype("Elf")
					return effects.SearchLibrary{
						Player:    item.Controller,
						Predicate: func(c game.Card) bool { return c.IsCreature() && c.HasSubtype(elf) },
						Dest:      game.ZoneBattlefield,
						Unbounded: true,
						Shuffle:   true,
						Source:    item.SourceCardID,
						Reason:    "Nissa Revane — any number of Elf creature cards",
					}.Apply(ctx)
				},
			},
		},
	})
}
