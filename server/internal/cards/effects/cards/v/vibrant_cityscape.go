package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vibrant Cityscape — Land (EDHREC rank 1454):
//
//	"{T}, Sacrifice this land: Search your library for a basic land
//	 card, put it onto the battlefield tapped, then shuffle."
//
// Evolving Wilds with a new name: the searcher picks the basic, it
// arrives tapped, the library shuffles. Its own file because the
// catalog is keyed on oracle ID, exactly as Terramorphic Expanse is.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6a7f3e1f-6798-4644-b64c-7765f81f0938",
		Name:         "Vibrant Cityscape",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Sacrifice this land: Search your library for a basic land card, put it onto the battlefield tapped, then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Lands: 1},
			Cost:    effects.Plus(effects.TapCost(), effects.SacrificeThis()),
			Effect:  effects.FetchBasicTapped,
		}},
	})
}
