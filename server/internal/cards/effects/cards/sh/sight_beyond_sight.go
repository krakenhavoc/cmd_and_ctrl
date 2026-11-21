package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sight Beyond Sight — Sorcery {3}{U}:
//
//	"Look at the top two cards of your library. Put one of them into
//	 your hand and the other on the bottom of your library.
//	 Rebound (If you cast this spell from your hand, exile it as it
//	 resolves. At the beginning of your next upkeep, you may cast this
//	 card from exile without paying its mana cost.)"
//
// Teferi, Temporal Archmage's +1 as a spell
// (LookAtTopThenTakeOneRestOnBottom). Rebound is the engine's keyword
// (game/rebound.go, #1854).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c2905e12-8af1-46f4-b888-be272a9b9748",
		Name:            "Sight Beyond Sight",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordRebound},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.LookAtTopThenTakeOneRestOnBottom(2,
				"Sight Beyond Sight — put one into your hand and the other on the bottom")(ctx.Game, item)
		},
	})
}
