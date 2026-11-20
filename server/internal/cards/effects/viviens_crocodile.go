package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vivien's Crocodile — Creature — Crocodile Spirit {2}{G}, 3/3 (Core Set 2020, #5518):
//
//	"This creature gets +1/+1 as long as you control a Vivien planeswalker."
//
// Briarbridge Tracker's conditional self pump (selfWhileYouControlA) over
// a planeswalker with the Vivien subtype.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b49791f9-525c-4b04-954d-d7678c5565e3",
		Name:         "Vivien's Crocodile",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "Gets +1/+1 as long as you control a Vivien planeswalker.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: selfWhileYouControlA(And(Planeswalker(), Subtype("Vivien"))),
			Apply:     pumpBy(1, 1),
		}},
	})
}
