package effects

// Norwood Archers — Creature — Elf Archer {3}{G}, 3/3 (Portal Second Age,
// #5625):
//
//	"Reach (This creature can block creatures with flying.)"
//
// Reach is an engine keyword carried as printed data: it can block a
// creature with flying (CR 702.17b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4e4e96aa-2f05-4f1d-96f8-6c42cd3be589",
		Name:            "Norwood Archers",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
	})
}
