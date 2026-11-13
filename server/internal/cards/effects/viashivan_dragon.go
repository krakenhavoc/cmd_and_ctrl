package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Viashivan Dragon — Creature — Dragon {2}{R}{R}{G}{G}, 4/4 (Visions, #5631):
//
//	"Flying
//	 {R}: This creature gets +1/+0 until end of turn.
//	 {G}: This creature gets +0/+1 until end of turn."
//
// Flying is an engine keyword carried as printed data (CR 702.9b). Each
// boost is a layer 7c change (CR 613.4c) that ends at cleanup
// (CR 514.2); neither has {T} in its cost, so both can be activated any
// number of times.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "86fb1fc1-7784-4c94-a80c-21ad87e50fc4",
		Name:            "Viashivan Dragon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{
			{
				Label:   "{R}: This creature gets +1/+0 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1}},
				Cost:    ManaCost("{R}"),
				Effect:  thisGetsUntilEndOfTurn(1, 0, "Viashivan Dragon — +1/+0 until end of turn"),
			},
			{
				Label:   "{G}: This creature gets +0/+1 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Toughness: 1}},
				Cost:    ManaCost("{G}"),
				Effect:  thisGetsUntilEndOfTurn(0, 1, "Viashivan Dragon — +0/+1 until end of turn"),
			},
		},
	})
}
