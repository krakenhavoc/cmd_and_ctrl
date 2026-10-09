package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rabbit Battery — Artifact Creature — Equipment Rabbit {R}, 1/1:
//
//	"Haste
//	 Equipped creature gets +1/+1 and has haste.
//	 Reconfigure {R}"
//
// Reconfigure is reconfigure.go (#2639).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c739e180-2f14-41ed-8e7e-50b7df985f35",
		Name:            "Rabbit Battery",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Static: []game.StaticAbility{
			PumpAttached(1, 1),
			GrantToAttached("haste"),
		},
		Activated: Reconfigure("{R}"),
	})
}
