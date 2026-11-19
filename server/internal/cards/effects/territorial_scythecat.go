package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Territorial Scythecat — Creature — Cat {2}{G}, 2/1 (Foundations Jumpstart):
//
//	"Trample
//	 Landfall — Whenever a land you control enters, put a +1/+1 counter on
//	 this creature."
//
// Floral Evoker's landfall counter; trample is an engine keyword.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "75d38954-6148-4d95-bf4b-cb82efcbe8df",
		Name:            "Territorial Scythecat",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			Landfall("Territorial Scythecat — put a +1/+1 counter on this creature", Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
