package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Excavation Mole — Creature — Mole {2}{G}, 3/3 (Ikoria: Lair of Behemoths, #5512):
//
//	"Trample
//	 When this creature enters, mill three cards."
//
// Aftermath Analyst's entry mill.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "68e0b691-f332-4b0a-8173-bb8a48b463a1",
		Name:            "Excavation Mole",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Excavation Mole — mill three cards", Do(MillCards{N: 3})),
		},
	})
}
