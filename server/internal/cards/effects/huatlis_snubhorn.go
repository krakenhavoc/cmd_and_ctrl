package effects

// Huatli's Snubhorn — Creature — Dinosaur {1}{W}, 2/2 (Ixalan, #5532):
//
//	"Vigilance"
//
// Vigilance is an engine keyword carried as printed data: attacking
// doesn't tap it (CR 702.20b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8516d8ba-b3f9-4c45-8d70-5dab476e478f",
		Name:            "Huatli's Snubhorn",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
	})
}
