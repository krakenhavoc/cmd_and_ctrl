package effects

// Darksteel Myr — Artifact Creature — Myr {3}, 0/1 (Scars of Mirrodin,
// #5577):
//
//	"Indestructible"
//
// A keyword body: indestructible is the engine's (CR 702.12b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f90fe86c-5cca-483b-93c2-6e856fa01c88",
		Name:            "Darksteel Myr",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
	})
}
