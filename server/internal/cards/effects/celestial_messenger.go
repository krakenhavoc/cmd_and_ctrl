package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Celestial Messenger — Creature — Bird Spirit {2}{U}{U}, 3/2 (Core Set 2020, #5518):
//
//	"Flash (You may cast this card any time you could cast an instant.)
//	 Flying
//	 This creature gets +1/+1 as long as you control a Yanling planeswalker."
//
// Briarbridge Tracker's conditional self pump (selfWhileYouControlA) over
// a planeswalker with the Yanling subtype.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "cf40caa2-cd5b-4372-ad4a-7f4cf564241e",
		Name:            "Celestial Messenger",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Static: []game.StaticAbility{{
			Label:     "Gets +1/+1 as long as you control a Yanling planeswalker.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: selfWhileYouControlA(And(Planeswalker(), Subtype("Yanling"))),
			Apply:     pumpBy(1, 1),
		}},
	})
}
