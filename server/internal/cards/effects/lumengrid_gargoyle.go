package effects

// Lumengrid Gargoyle — Artifact Creature — Gargoyle {6}, 4/4 (Mirrodin Besieged,
// #5576):
//
//	"Flying"
//
// Flying is a printed keyword read at the block declaration (CR 702.9b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1070ede7-c81d-4c77-a1a9-6f243432a640",
		Name:            "Lumengrid Gargoyle",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
