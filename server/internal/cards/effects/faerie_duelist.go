package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Faerie Duelist — Creature — Faerie Rogue {1}{U}, 1/2 (Ravnica Allegiance,
// #5521):
//
//	"Flash
//	 Flying
//	 When this creature enters, target creature an opponent controls gets
//	 -2/-0 until end of turn."
//
// Flash and flying are printed data (CR 702.8a, 702.9b). The enters
// trigger is Humbling Elder's -2/-0 on a creature an opponent controls.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ccbf7141-69c9-407d-b68e-c1c63ddef213",
		Name:            "Faerie Duelist",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(WhenThisEnters("Faerie Duelist — target creature an opponent controls gets -2/-0 until end of turn",
				boostTargetUntilEOT("Faerie Duelist — -2/-0 until end of turn", -2, 0)),
				TargetCreature("target creature an opponent controls", OpponentControls())),
				ForTargets(HarmUntilEOT(0, -2, 0))),
		},
	})
}
