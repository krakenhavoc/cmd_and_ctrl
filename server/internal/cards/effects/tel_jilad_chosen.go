package effects

// Tel-Jilad Chosen — Creature — Elf Warrior {1}{G}, 2/1 (Mirrodin, #5606):
//
//	"Protection from artifacts"
//
// Nacatl Savage's printed keyword: protection from artifacts, read by
// targeting, blocking, damage and attaching (CR 702.16a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "de6854d1-bdb0-4b9a-b442-08d4d90a538d",
		Name:            "Tel-Jilad Chosen",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"protection from artifacts"},
	})
}
