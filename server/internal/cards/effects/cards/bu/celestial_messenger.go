package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "cf40caa2-cd5b-4372-ad4a-7f4cf564241e",
		Name:            "Celestial Messenger",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Static: []game.StaticAbility{{
			Label:     "Gets +1/+1 as long as you control a Yanling planeswalker.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: effects.SelfWhileYouControlA(effects.And(effects.Planeswalker(), effects.Subtype("Yanling"))),
			Apply:     effects.PumpBy(1, 1),
		}},
	})
}
