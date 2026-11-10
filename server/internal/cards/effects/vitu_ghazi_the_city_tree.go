package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vitu-Ghazi, the City-Tree — Land:
//
//	"{T}: Add {C}.
//	 {2}{G}{W}, {T}: Create a 1/1 green Saproling creature token."
//
// Castle Ardenvale's token maker with a Saproling.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "88fb9e82-28b9-4275-a1e2-cb3a9bfda127",
		Name:          "Vitu-Ghazi, the City-Tree",
		Completeness:  CompletenessFull,
		ManaAbilities: []ManaAbility{cmdr2B01TapForColorless()},
		Activated: []ActivatedAbility{{
			Label:   "{2}{G}{W}, {T}: Create a 1/1 green Saproling creature token.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    Plus(ManaCost("{2}{G}{W}"), TapCost()),
			Effect:  cmdr2B01MakeToken("1/1 green Saproling"),
		}},
	})
}
