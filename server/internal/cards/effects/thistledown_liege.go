package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thistledown Liege — Creature — Kithkin Knight {1}{W/U}{W/U}{W/U}, 1/3
// (Shadowmoor, #5591):
//
//	"Flash
//	 Other white creatures you control get +1/+1.
//	 Other blue creatures you control get +1/+1."
//
// Flash is a printed keyword. Glen Elendra Liege's two colour anthems:
// each printed ability applies on its own, so a creature of both colours
// gets +2/+2, and "other" excludes the Liege by instance.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7b07356f-9c3f-4875-939e-aa64eb6b5ce2",
		Name:            "Thistledown Liege",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Static: []game.StaticAbility{
			b16Anthem(g2OtherCreaturesYouControlOfColor("W"), 1, 1),
			b16Anthem(g2OtherCreaturesYouControlOfColor("U"), 1, 1),
		},
	})
}
