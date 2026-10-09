package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lizard Blades — Artifact Creature — Equipment Lizard {1}{R}, 1/1:
//
//	"Double strike
//	 Equipped creature has double strike.
//	 Reconfigure {2}"
//
// Reconfigure is reconfigure.go (#2639). While it is attached it is not
// a creature (CR 702.151b), so its own double strike matters only while
// it fights on its own.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f6314e50-a610-4fff-8f60-803d23460ff0",
		Name:            "Lizard Blades",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"double strike"},
		Static:          []game.StaticAbility{GrantToAttached("double strike")},
		Activated:       Reconfigure("{2}"),
	})
}
