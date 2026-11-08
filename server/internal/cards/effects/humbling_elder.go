package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Humbling Elder — Creature — Human Monk {U}, 1/2 (Tarkir: Dragonstorm,
// #4463):
//
//	"Flash
//	 When this creature enters, target creature an opponent controls gets
//	 -2/-0 until end of turn."
//
// Cogwork Wrestler's text on a Monk. The trigger is removed with no legal
// target (CR 603.3d), and the shrink ends in the cleanup step (CR 514.2).
// A target no longer legal is left alone (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "013bcab8-e527-4700-82c9-4cb9cb8a4807",
		Name:            "Humbling Elder",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(WhenThisEnters("Humbling Elder — target creature an opponent controls gets -2/-0 until end of turn",
				boostTargetUntilEOT("Humbling Elder — -2/-0 until end of turn", -2, 0)),
				TargetCreature("target creature an opponent controls", OpponentControls())),
				ForTargets(HarmUntilEOT(0, -2, 0))),
		},
	})
}
