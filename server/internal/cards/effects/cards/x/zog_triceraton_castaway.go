package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "d944d33d-0a95-4d0d-96ce-49d02500227d",
		Name:            "Zog, Triceraton Castaway",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach", "trample"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters(label, effects.TargetCreatureCantBlockThisTurn(label)),
				effects.TargetCreature("target creature")),
		},
		Activated: []effects.ActivatedAbility{
			effects.Typecycling("Mountaincycling", "{2}", "a Mountain card", effects.IsLandWithSubtype("Mountain")),
		},
	})
}
