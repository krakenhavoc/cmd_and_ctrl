package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Terminal Moraine — Land (Dominaria Remastered, #5948):
//
//	"{T}: Add {C}.
//	 {2}, {T}, Sacrifice this land: Search your library for a basic land
//	 card, put that card onto the battlefield tapped, then shuffle."
//
// Evolving Wilds' fetch behind a {2} cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "58eaaa8b-45c6-439b-bdd1-5f4e77a75a8c",
		Name:          "Terminal Moraine",
		Completeness:  effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{effects.PainlessColorless()},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}, Sacrifice this land: Search your library for a basic land card, put that card onto the battlefield tapped, then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Lands: 1},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost(), effects.SacrificeThis()),
			Effect:  effects.FetchBasicTapped,
		}},
	})
}
