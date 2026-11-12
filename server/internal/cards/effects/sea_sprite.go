package effects

// Sea Sprite — Creature — Faerie {1}{U}, 1/1 (Homelands, #5635):
//
//	"Flying, protection from red"
//
// Both are engine keywords carried as printed data: flying limits its
// blockers (CR 702.9b), and protection from red stops red spells and
// abilities from red sources targeting it and prevents damage from red
// sources (CR 702.16b, CR 702.16e).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b318d94e-314f-4df8-a178-b5ddd756e1c4",
		Name:            "Sea Sprite",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "protection from red"},
	})
}
