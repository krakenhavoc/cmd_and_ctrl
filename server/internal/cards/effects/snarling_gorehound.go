package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Snarling Gorehound — Creature — Dog {B}, 1/1 (Murders at Karlov
// Manor, #4469):
//
//	"Menace
//	 Whenever another creature you control with power 2 or less enters,
//	 surveil 1. (Look at the top card of your library. You may put it
//	 into your graveyard.)"
//
// Marketwatch Phantom's trigger condition (Mentor of the Meek's, the
// power read as the creature enters), with the ordinary surveil
// (CR 701.25a) as its effect. The surveil does not depend on the
// Gorehound still being on the battlefield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d2b1c5eb-06f0-4ec4-9446-791faac099a7",
		Name:            "Snarling Gorehound",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, AnotherCreatureWithPower2OrLessEnteredUnderYourControl,
				"Snarling Gorehound — surveil 1", Do(Surveil{N: 1})),
		},
	})
}
