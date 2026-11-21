package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Terramorph — Sorcery {3}{G}:
//
//	"Search your library for a basic land card, put it onto the
//	 battlefield, then shuffle.
//	 Rebound (If you cast this spell from your hand, exile it as it
//	 resolves. At the beginning of your next upkeep, you may cast this
//	 card from exile without paying its mana cost.)"
//
// The land enters untapped: the card says nothing about tapped.
// Rebound is the engine's keyword (game/rebound.go, #1854).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "17a34f8d-a80f-4331-8be5-06cbb9d10d7b",
		Name:            "Terramorph",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordRebound},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.SearchLibrary{
				Player:    ctx.Controller(),
				Predicate: effects.IsBasicLand,
				Dest:      game.ZoneBattlefield,
				Limit:     1,
				Reveal:    true,
				Shuffle:   true,
				Reason:    "Terramorph — a basic land",
			}.Apply(ctx)
		},
	})
}
