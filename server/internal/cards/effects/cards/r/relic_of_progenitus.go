package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Relic of Progenitus — Artifact {1}:
//
//	"{T}: Target player exiles a card from their graveyard.
//	 {1}, Exile this artifact: Exile all graveyards. Draw a card."
//
// The first ability targets a player, and THAT player chooses the card,
// from their own graveyard as the ability resolves (an empty graveyard
// exiles nothing). The second pays {1} and exiles the Relic as a cost,
// then exiles every seat's graveyard (yours too) and draws.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6a9c3401-570c-41cb-a605-75dfa57d6ab7",
		Name:         "Relic of Progenitus",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{T}: Target player exiles a card from their graveyard.",
				Purpose: game.Purpose{Answers: game.AnswerGraveyard},
				Cost:    effects.TapCost(),
				Targets: effects.TargetPlayer("target player"),
				Effect:  effects.RprB15TargetExilesOwnGraveyardCard,
			},
			{
				Label:   "{1}, Exile this artifact: Exile all graveyards. Draw a card.",
				Purpose: game.Purpose{Answers: game.AnswerGraveyard, Draws: 1},
				Cost:    effects.Plus(effects.ManaCost("{1}"), effects.ExileThis()),
				Effect:  effects.RprB15ExileAllGraveyardsThenDraw,
			},
		},
	})
}
