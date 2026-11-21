package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vitu-Ghazi, the City-Tree — Land:
//
//	"{T}: Add {C}.
//	 {2}{G}{W}, {T}: Create a 1/1 green Saproling creature token."
//
// Castle Ardenvale's token maker with a Saproling.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "88fb9e82-28b9-4275-a1e2-cb3a9bfda127",
		Name:          "Vitu-Ghazi, the City-Tree",
		Completeness:  effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{effects.PainlessColorless()},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{G}{W}, {T}: Create a 1/1 green Saproling creature token.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    effects.Plus(effects.ManaCost("{2}{G}{W}"), effects.TapCost()),
			Effect:  effects.Cmdr2B01MakeToken("1/1 green Saproling"),
		}},
	})
}
