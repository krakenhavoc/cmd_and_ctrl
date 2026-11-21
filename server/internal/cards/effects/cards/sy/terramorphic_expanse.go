package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Terramorphic Expanse — Land (EDHREC rank 28):
//
//	"{T}, Sacrifice this land: Search your library for a basic land
//	card, put it onto the battlefield tapped, then shuffle."
//
// Evolving Wilds under a different name. Both are in the catalog
// because both are in decks — a list that runs one very often runs
// the other as the eleventh copy.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1bd3e453-aa21-4ee6-95c2-d6d920ee8e7a",
		Name:         "Terramorphic Expanse",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Sacrifice this land: Search your library for a basic land card, put it onto the battlefield tapped, then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Lands: 1},
			Cost:    effects.Plus(effects.TapCost(), effects.SacrificeThis()),
			Effect:  effects.FetchBasicTapped,
		}},
	})
}
