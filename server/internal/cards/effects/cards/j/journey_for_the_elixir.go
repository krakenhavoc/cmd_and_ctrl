package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Journey for the Elixir (Global Series Jiang Yanggu & Mu Yanling, #5528):
//
//	Search your library and graveyard for a basic land card and a card named Jiang Yanggu, reveal them, put them into your hand, then shuffle.
//
// {2}{G} Sorcery. Both zones are searched (BothZones, #6156); the set rule lets
// the searcher take at most one basic land card and at most one card named
// Jiang Yanggu, and either may fail to be found.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e87606d5-5a23-4507-aece-8fdd044e7185",
		Name:         "Journey for the Elixir",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 2},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return ctx.Game.SearchLibraryThenForEffect(game.SearchLibrarySpec{
				Player: ctx.Controller(),
				Source: ctx.Source(),
				Pred: func(c game.Card) bool {
					return effects.S82OldB203BasicLandOrJiang(c)
				},
				Dest:          game.ZoneHand,
				Limit:         2,
				Reveal:        true,
				Shuffle:       true,
				AlsoGraveyard: true,
				BothZones:     true,
				Rule:          effects.SearchRuleBasicLandAndJiangYanggu,
				Reason:        "Journey for the Elixir — a basic land card and a card named Jiang Yanggu",
			})
		},
	})
}
