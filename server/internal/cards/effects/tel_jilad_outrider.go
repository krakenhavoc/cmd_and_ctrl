package effects

// Tel-Jilad Outrider — Creature — Elf Warrior {3}{G} (Darksteel, #5605):
//
//	"Protection from artifacts"
//
// Nacatl Savage's printed protection (CR 702.16).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d83d784e-e95b-48a1-9b4a-362f4fccc529",
		Name:            "Tel-Jilad Outrider",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"protection from artifacts"},
	})
}
