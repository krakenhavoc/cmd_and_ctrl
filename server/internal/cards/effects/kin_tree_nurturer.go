package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kin-Tree Nurturer — Creature — Human Druid {2}{B}, 2/1 (Tarkir:
// Dragonstorm, #4463):
//
//	"Lifelink
//	 When this creature enters, it endures 1. (Put a +1/+1 counter on it
//	 or create a 1/1 white Spirit creature token.)"
//
// Endure is endure_n.go (CR 701.63a), Sandskitter Outrider's entry
// trigger. The Spirit does not have lifelink: only the Nurturer does.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ad28a1f9-8893-4693-9464-c32ed760f6b1",
		Name:            "Kin-Tree Nurturer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Kin-Tree Nurturer — it endures 1", ThisEndures("Kin-Tree Nurturer", 1)),
		},
	})
}
