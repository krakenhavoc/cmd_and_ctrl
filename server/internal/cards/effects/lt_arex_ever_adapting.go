package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lt. Arex, Ever-Adapting — Legendary Creature — Edosian Officer
// {5}{G}{U}, */* (Star Trek Commander, #4974):
//
//	"Trample
//	 Strange new worlds — Lt. Arex's power and toughness are each equal
//	 to the number of differently named lands you control.
//	 When Lt. Arex enters, draw three cards. You may put up to three land
//	 cards from your hand onto the battlefield tapped."
//
// "Strange new worlds" is an ability word (CR 207.2c); the P/T is Captain
// Pike, Fullness of Time's characteristic-defining ability (CR 604.3a),
// counted in layer 7a, each land name once. The enters trigger draws
// three, then offers up to three land cards from the hand, any of the
// three just drawn included; they enter tapped, together, and the put
// waits for any question the draw asks. Declining puts nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "51a2671b-677e-434c-9f1a-d768fe1190c9",
		Name:            "Lt. Arex, Ever-Adapting",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Purpose:         game.Purpose{Draws: 3},
		Static: []game.StaticAbility{
			trekB19DifferentlyNamedLandsPT("Strange new worlds — Lt. Arex's power and toughness are each equal to the number of differently named lands you control."),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Lt. Arex, Ever-Adapting — draw three cards, then you may put up to three land cards from your hand onto the battlefield tapped",
				Do(
					DrawCards{N: 3},
					PutFromHandOntoBattlefield{
						Match:    Land(),
						Tapped:   true,
						Optional: true,
						Max:      3,
						Label:    "Lt. Arex, Ever-Adapting — put up to three land cards from your hand onto the battlefield tapped",
					},
				)),
		},
	})
}
