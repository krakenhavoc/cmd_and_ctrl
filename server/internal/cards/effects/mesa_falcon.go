package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mesa Falcon — Creature — Bird {1}{W}, 1/1 (Homelands, #5635):
//
//	"Flying
//	 {1}{W}: This creature gets +0/+1 until end of turn."
//
// Flying is an engine keyword carried as printed data (CR 702.9b). The
// boost is a layer 7c change (CR 613.4c) that ends at cleanup
// (CR 514.2); with no {T} in the cost it can be activated any number of
// times, including the turn the Falcon arrives.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "bc4c5f33-0f3d-4815-9974-fb12ac725d08",
		Name:            "Mesa Falcon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:   "{1}{W}: This creature gets +0/+1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Toughness: 1}},
			Cost:    ManaCost("{1}{W}"),
			Effect:  thisGetsUntilEndOfTurn(0, 1, "Mesa Falcon — +0/+1 until end of turn"),
		}},
	})
}
