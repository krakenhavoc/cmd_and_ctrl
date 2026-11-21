package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Boartusk Liege — Creature — Goblin Knight {1}{R/G}{R/G}{R/G}, 3/4
// (Shadowmoor, #5591):
//
//	"Trample
//	 Other red creatures you control get +1/+1.
//	 Other green creatures you control get +1/+1."
//
// Trample is a printed keyword. Glen Elendra Liege's two colour anthems:
// each printed ability applies on its own, so a creature of both colours
// gets +2/+2, and "other" excludes the Liege by instance.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4fcffb56-5ab6-44e2-804e-8b3c55cc5415",
		Name:            "Boartusk Liege",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Static: []game.StaticAbility{
			effects.B16Anthem(effects.G2OtherCreaturesYouControlOfColor("R"), 1, 1),
			effects.B16Anthem(effects.G2OtherCreaturesYouControlOfColor("G"), 1, 1),
		},
	})
}
