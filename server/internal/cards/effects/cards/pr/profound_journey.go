package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Profound Journey — Sorcery {5}{W}{W}:
//
//	"Return target permanent card from your graveyard to the
//	 battlefield.
//	 Rebound (If you cast this spell from your hand, exile it as it
//	 resolves. At the beginning of your next upkeep, you may cast this
//	 card from exile without paying its mana cost.)"
//
// The card returns under its owner's control, which is you: it came
// from your graveyard. Rebound is the engine's keyword
// (game/rebound.go, #1854).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "070e3224-0f89-4716-b91b-0131eff6146f",
		Name:            "Profound Journey",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordRebound},
		Targets:         effects.TargetCardInGraveyard("target permanent card in your graveyard", effects.YouOwn(), effects.Permanent()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.ReturnFirstLegalGraveyardTargetToBattlefield(ctx.Game, item)
		},
	})
}
