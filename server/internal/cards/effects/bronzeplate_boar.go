package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bronzeplate Boar — Artifact Creature — Equipment Boar {2}{R}, 3/2:
//
//	"Trample
//	 Equipped creature gets +3/+2 and has trample.
//	 Reconfigure {5}"
//
// Reconfigure is reconfigure.go (#2639).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0a09f825-0484-40b4-8adf-1028385db480",
		Name:            "Bronzeplate Boar",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Static: []game.StaticAbility{
			PumpAttached(3, 2),
			GrantToAttached("trample"),
		},
		Activated: Reconfigure("{5}"),
	})
}
