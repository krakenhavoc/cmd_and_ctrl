package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dirtwater Wraith — Creature — Wraith {3}{B}, 1/3 (Mirage, #5633):
//
//	"Swampwalk (This creature can't be blocked as long as defending
//	 player controls a Swamp.)
//	 {B}: This creature gets +1/+0 until end of turn."
//
// Swampwalk is a printed keyword the block legality check reads
// (CR 702.14c). The pump is the shared thisGetsUntilEndOfTurn body.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0126f620-1dbf-405c-b5cf-f48b7ef02d32",
		Name:            "Dirtwater Wraith",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"swampwalk"},
		Activated: []ActivatedAbility{{
			Label:   "{B}: This creature gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1}},
			Cost:    ManaCost("{B}"),
			Effect:  thisGetsUntilEndOfTurn(1, 0, "Dirtwater Wraith — +1/+0 until end of turn"),
		}},
	})
}
