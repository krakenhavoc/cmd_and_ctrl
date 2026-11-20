package effects

// Swift Spinner — Creature — Spider {3}{G}, 2/3 (Eldritch Moon, #5541):
//
//	"Flash (You may cast this spell any time you could cast an instant.)
//	 Reach (This creature can block creatures with flying.)"
//
// Dawn's Light Archer's two keywords, carried as printed data.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "91585b64-f10d-47c6-a563-384ceb944710",
		Name:            "Swift Spinner",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "reach"},
	})
}
