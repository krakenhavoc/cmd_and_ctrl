package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Loam Lion — Creature — Cat {W}, 1/1 (Worldwake, #5582):
//
//	"This creature gets +1/+2 as long as you control a Forest."
//
// Kird Ape's static: a layer-7c pump while its controller controls a
// Forest.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "18961e25-3feb-4dba-a77f-240a4d664ecb",
		Name:         "Loam Lion",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "This creature gets +1/+2 as long as you control a Forest.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: selfWhileYouControlA(Subtype("Forest")),
			Apply:     pumpBy(1, 2),
		}},
	})
}
