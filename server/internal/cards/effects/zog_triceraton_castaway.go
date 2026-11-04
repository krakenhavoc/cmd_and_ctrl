package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Zog, Triceraton Castaway — {4}{R} Legendary Creature — Dinosaur Soldier
// 5/4 (Teenage Mutant Ninja Turtles, #4457):
//
//	"Reach, trample
//	 When Zog enters, target creature can't block this turn.
//	 Mountaincycling {2}"
//
// The enters trigger is Ahn-Crop Crasher's body
// (targetCreatureCantBlockThisTurn); mountaincycling is the engine's
// typecycling from the hand (CR 702.29e), as on Oliphaunt.
//
// No simplification.
func init() {
	const label = "Zog — target creature can't block this turn"
	Register(Spec{
		OracleID:        "d944d33d-0a95-4d0d-96ce-49d02500227d",
		Name:            "Zog, Triceraton Castaway",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", "trample"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters(label, targetCreatureCantBlockThisTurn(label)),
				TargetCreature("target creature")),
		},
		Activated: []ActivatedAbility{
			Typecycling("Mountaincycling", "{2}", "a Mountain card", IsLandWithSubtype("Mountain")),
		},
	})
}
