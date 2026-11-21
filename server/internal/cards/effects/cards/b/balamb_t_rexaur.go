package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "3f678f76-3863-4ab5-a4fd-9b76a282b1a8",
		Name:            "Balamb T-Rexaur",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Balamb T-Rexaur — you gain 3 life", effects.Do(effects.GainLife{Amount: 3})),
		},
		Activated: []effects.ActivatedAbility{
			effects.Typecycling("Forestcycling", "{2}", "a Forest card", effects.IsLandWithSubtype("Forest")),
		},
	})
}
