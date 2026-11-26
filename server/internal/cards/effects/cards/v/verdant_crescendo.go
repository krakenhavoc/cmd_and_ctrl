package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Verdant Crescendo — Sorcery {3}{G} (Kaladesh, #5539):
//
//	"Search your library for a basic land card and put it onto the
//	 battlefield tapped. Search your library and graveyard for a card named
//	 Nissa, Nature's Artisan, reveal it, put it into your hand, then
//	 shuffle."
//
// Cultivate's two chained searches: the land search defers its shuffle and
// hands on to the second, which searches both zones (BothZones: "library
// and graveyard", not "and/or"), reveals the planeswalker and shuffles once,
// as printed. Neither search is optional, and either may find nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b184d12d-41d8-4a74-99b1-8c7fced212f8",
		Name:         "Verdant Crescendo",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Lands: 1, Tutors: 1},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.SearchLibrary{
				Player:        ctx.Controller(),
				Source:        ctx.Source(),
				Predicate:     effects.IsBasicLand,
				Dest:          game.ZoneBattlefield,
				Limit:         1,
				TappedOnEntry: true,
				Reason:        "Verdant Crescendo — a basic land card, onto the battlefield tapped",
				ThenTail:      ctx.ItemTail(verdantCrescendoNissa),
			}.Apply(ctx)
		},
	})
}

// verdantCrescendoNissa is the second sentence: search the library and the
// graveyard for Nissa, reveal her, put her into the hand, then shuffle. An
// on-disk key.
var verdantCrescendoNissa = effects.CardTail("tail/verdant-crescendo-nissa", func(ctx *effects.Context, _ game.TailOutcome) error {
	return ctx.Game.SearchLibraryThenForEffect(game.SearchLibrarySpec{
		Player:        ctx.Controller(),
		Source:        ctx.Source(),
		Pred:          func(c game.Card) bool { return c.Name == "Nissa, Nature's Artisan" },
		Dest:          game.ZoneHand,
		Limit:         1,
		Reveal:        true,
		Shuffle:       true,
		AlsoGraveyard: true,
		BothZones:     true,
		Reason:        "Verdant Crescendo — a card named Nissa, Nature's Artisan",
	})
})
