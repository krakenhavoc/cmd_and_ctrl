package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Burrog Befuddler — Creature — Frog Wizard {1}{U}, 2/1 (Foundations,
// #4465):
//
//	"Flash
//	 When this creature enters, target creature an opponent controls
//	 gets -1/-0 until end of turn."
//
// Flash rides PrintedKeywords; the enter trigger shrinks the
// still-legal target (boostTargetUntilEOT with a negative power).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "73b60c2c-7c74-4554-8861-daa0d6fc22b4",
		Name:            "Burrog Befuddler",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Burrog Befuddler — target creature an opponent controls gets -1/-0 until end of turn",
				boostTargetUntilEOT("Burrog Befuddler — -1/-0", -1, 0)),
				TargetCreature("target creature an opponent controls", OpponentControls())),
		},
	})
}
