package effects

// Pilgrim of the Fires — Artifact Creature — Golem {7}, 6/4 (Fate Reforged,
// #5552):
//
//	"First strike, trample"
//
// Every word of it is an engine keyword carried as printed data.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "53390f91-a159-482f-9f64-f5adc812680b",
		Name:            "Pilgrim of the Fires",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike", "trample"},
	})
}
