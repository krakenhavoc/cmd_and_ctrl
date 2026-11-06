package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Balamb T-Rexaur — Creature — Dinosaur {4}{G}{G}, 6/6 (Final Fantasy,
// #4462):
//
//	"Trample
//	 When this creature enters, you gain 3 life.
//	 Forestcycling {2}"
//
// The enters trigger uses the stack. Forestcycling is the shared
// typecycling ability from the hand (Rocksteady, Crash Courser's): any
// card with the Forest land type, revealed, then the library is
// shuffled.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3f678f76-3863-4ab5-a4fd-9b76a282b1a8",
		Name:            "Balamb T-Rexaur",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Balamb T-Rexaur — you gain 3 life", Do(GainLife{Amount: 3})),
		},
		Activated: []ActivatedAbility{
			Typecycling("Forestcycling", "{2}", "a Forest card", IsLandWithSubtype("Forest")),
		},
	})
}
