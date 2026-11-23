package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Void Squall — Sorcery {4}{U}:
//
//	"Return target nonland permanent to its owner's hand.
//	 Rebound (If you cast this spell from your hand, exile it as it
//	 resolves. At the beginning of your next upkeep, you may cast this
//	 card from exile without paying its mana cost.)"
//
// Rebound is the engine's keyword (game/rebound.go, #1854).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e8ca5c7e-7d8a-4e86-91ee-2d3576f9fd5d",
		Name:            "Void Squall",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordRebound},
		Targets:         effects.TargetPermanent("target nonland permanent", effects.Nonland()),
		OnResolve:       effects.BounceTheTarget,
	})
}
