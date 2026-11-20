package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Khenra Charioteer — Creature — Jackal Warrior {1}{R}{G}, 3/3
// (Amonkhet, #5535):
//
//	"Trample
//	 Other creatures you control have trample."
//
// Aggressive Mammoth's grant.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e75bc5df-f4a4-45af-a6fd-2711b98dd521",
		Name:            "Khenra Charioteer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Static: []game.StaticAbility{
			TribalKeywordGrant(TribeFilter{Others: true, YoursOnly: true}, "trample"),
		},
	})
}
