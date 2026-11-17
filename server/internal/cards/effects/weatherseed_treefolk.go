package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Weatherseed Treefolk — Creature — Treefolk {2}{G}{G}{G}, 5/3 (Urza's
// Legacy, #5623):
//
//	"Trample
//	 When this creature dies, return it to its owner's hand."
//
// Immortal Phoenix's death trigger on a trampler, through the shared
// ulgB1DiesReturnToHand row: only a death (CR 700.4) sets it off, and it
// goes to its OWNER's hand while it is still the graveyard object that died
// (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4e9b878b-80a2-433b-941b-92d854e83d58",
		Name:            "Weatherseed Treefolk",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			ulgB1DiesReturnToHand("Weatherseed Treefolk"),
		},
	})
}
