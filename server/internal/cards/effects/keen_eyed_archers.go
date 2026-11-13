package effects

// Keen-Eyed Archers — Creature — Elf Archer {2}{W}, 2/2 (Portal, #5630):
//
//	"Reach"
//
// Reach is an engine keyword carried as printed data: it can block
// creatures with flying (CR 702.17b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0ace32d6-7261-447c-9ee2-e03febaab91b",
		Name:            "Keen-Eyed Archers",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
	})
}
