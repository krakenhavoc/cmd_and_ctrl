package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Soldevi Digger — Artifact {2} (Alliances, #5634):
//
//	"{2}: Put the top card of your graveyard on the bottom of your library."
//
// The top card is the one most recently put there (CR 404.2, #6194). It goes to
// the bottom of its owner's library; an empty graveyard does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b526962e-be8a-4803-b038-57869bb96751",
		Name:         "Soldevi Digger",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}: Put the top card of your graveyard on the bottom of your library.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{2}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				top := g.TopGraveyardCardsForEffect(item.Controller, 1, nil)
				if len(top) == 0 {
					return nil
				}
				return g.TuckCardsToLibraryKeyedForEffect(top, game.TuckOptions{ToBottom: true}, game.KeyedTail{})
			},
		}},
	})
}
