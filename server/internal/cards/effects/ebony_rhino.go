package effects

// Ebony Rhino — Artifact Creature — Rhino {7}, 4/5 (Homelands, #5635):
//
//	"Trample"
//
// Trample is the printed keyword the combat-damage assignment reads
// (CR 702.19b); the cost, the types and the 4/5 ride the card from
// Scryfall.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3d2b5201-c2bf-4753-89ad-543795740216",
		Name:            "Ebony Rhino",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
	})
}
