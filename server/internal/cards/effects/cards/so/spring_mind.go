package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Spring // Mind — Sorcery // Instant, Aftermath (Amonkhet, S82 OLD batch 243, #5535):
//
//	Spring {2}{G}: "Search your library for a basic land card, put it onto the
//	 battlefield tapped, then shuffle."
//	Mind {4}{U}{U}: "Aftermath. Draw two cards."
//
// Two registered faces (the left under the oracle ID, the right under "#1").
// Mind is an instant, castable only from a graveyard and exiled as it resolves
// (CR 702.127a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6dd7e861-7ae2-49e6-a9ab-8fd761939e22",
		Name:         "Spring",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Lands: 1},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.SearchLibrary{
				Player:        item.Controller,
				Predicate:     effects.IsBasicLand,
				Dest:          game.ZoneBattlefield,
				Limit:         1,
				Reveal:        true,
				Shuffle:       true,
				TappedOnEntry: true,
				Reason:        "Spring — a basic land card, onto the battlefield tapped",
			}.Apply(ctx)
		},
	})
	effects.Register(effects.Spec{
		OracleID:     "6dd7e861-7ae2-49e6-a9ab-8fd761939e22#1",
		Name:         "Mind",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 2},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DrawCards{Player: ctx.Controller(), N: 2}.Apply(ctx)
		},
	})
}
