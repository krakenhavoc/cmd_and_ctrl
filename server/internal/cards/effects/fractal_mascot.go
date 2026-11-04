package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fractal Mascot — Creature — Fractal Elk {4}{G}{U}, 6/6 (Secrets of
// Strixhaven, #4456):
//
//	"Trample
//	 When this creature enters, tap target creature an opponent controls. Put
//	 a stun counter on it. (If a permanent with a stun counter would become
//	 untapped, remove one from it instead.)"
//
// The enters trigger is the shared tap-and-stun body (tapTargetAndStunIt,
// Deluge Virtuoso's): a target that has left or stopped being a legal pick by
// resolution is left alone (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "cca3f3e8-15a7-4ad8-893a-d3570178a800",
		Name:            "Fractal Mascot",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Fractal Mascot — tap target creature an opponent controls and put a stun counter on it", tapTargetAndStunIt),
				TargetCreature("target creature an opponent controls", OpponentControls())),
		},
	})
}
