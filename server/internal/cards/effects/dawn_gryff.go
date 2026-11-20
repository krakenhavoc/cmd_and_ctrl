package effects

// Dawn Gryff — Creature — Hippogriff {2}{W}, 2/2 (Eldritch Moon, #5541):
//
//	"Flying"
//
// Snapping Drake's keyword, carried as printed data.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1c34b2e4-e600-4688-86cb-8a1c11fdc770",
		Name:            "Dawn Gryff",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
