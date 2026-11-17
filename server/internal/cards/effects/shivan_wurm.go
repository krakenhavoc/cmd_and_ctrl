package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shivan Wurm — Creature — Wurm {3}{R}{G}, 7/7 (Planeshift, #5615):
//
//	"Trample
//	 When this creature enters, return a red or green creature you
//	 control to its owner's hand."
//
// Trample is a printed keyword; the enters trigger is the shared gating
// return (gating.go), chosen as it resolves, the Wurm itself included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "fc988ce3-0cad-41a5-8f13-656caca732df",
		Name:            "Shivan Wurm",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered:       []game.TriggeredAbility{gatingReturn("Shivan Wurm", "R", "G")},
	})
}
