package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Briarbridge Tracker — Creature — Human Scout {2}{G}, 2/3 (Midnight Hunt,
// #5497):
//
//	"Vigilance
//	 When this creature enters, investigate.
//	 As long as you control a token, this creature gets +2/+0."
//
// Druid of the Spade's static without the trample grant: "you control a
// token" reads the battlefield (selfWhileYouControlA over IsTokenPredicate),
// so the Clue it makes turns the bonus on at once.
//
// No simplification.
func init() {
	whileToken := selfWhileYouControlA(IsTokenPredicate())
	Register(Spec{
		OracleID:        "8fe6e954-f19b-4d11-83d3-74c634d51bad",
		Name:            "Briarbridge Tracker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, Self, "Briarbridge Tracker — investigate", Do(Investigate{N: 1})),
		},
		Static: []game.StaticAbility{{
			Label:     "Gets +2/+0 as long as you control a token.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: whileToken,
			Apply:     pumpBy(2, 0),
		}},
	})
}
