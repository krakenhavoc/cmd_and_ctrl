package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Flora Colossus — Creature — Treefolk {5}{G}{G}, */* (Marvel Super
// Heroes Commander, #4455):
//
//	"Hexproof
//	 Flora Colossus's power and toughness are each equal to the number
//	 of lands you control."
//
// A characteristic-defining ability (CR 604.3, layer 7a) counting lands
// live, the way Ashaya does (landsYouControlPT).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "04dce348-be42-443a-a88d-c47436ba2ec4",
		Name:            "Flora Colossus",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"hexproof"},
		Static: []game.StaticAbility{
			landsYouControlPT("Flora Colossus's power and toughness are each equal to the number of lands you control."),
		},
	})
}
