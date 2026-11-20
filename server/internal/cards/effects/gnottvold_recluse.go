package effects

// Gnottvold Recluse — Creature — Spider {2}{G}, 4/2 (Kaldheim, #5505):
//
//	"Reach"
//
// A keyword body: the keywords ride PrintedKeywords and the engine's
// combat and damage rules do the rest.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e7e65cc5-9ceb-43b6-acde-ae492b71daf6",
		Name:            "Gnottvold Recluse",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
	})
}
