package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Warped Landscape — Land (Jumpstart 2022, #5948):
//
//	"{T}: Add {C}.
//	 {2}, {T}, Sacrifice this land: Search your library for a basic land
//	 card, put it onto the battlefield tapped, then shuffle."
//
// Terramorphic Expanse's fetch (the shared FetchBasicTapped body) behind a
// {2} payment.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "543e6bb3-a867-43bf-a737-2f5d6d8dc631",
		Name:         "Warped Landscape",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}, Sacrifice this land: Search your library for a basic land card, put it onto the battlefield tapped, then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Lands: 1},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost(), effects.SacrificeThis()),
			Effect:  effects.FetchBasicTapped,
		}},
	})
}
