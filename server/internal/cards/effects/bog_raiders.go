package effects

// Bog Raiders — Creature — Zombie {2}{B}, 2/2 (Portal, #5630):
//
//	"Swampwalk"
//
// Swampwalk is an engine keyword carried as printed data: it can't be
// blocked while the defending player controls a Swamp (CR 702.14c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "60824fae-20ed-4122-82c9-e99a1b679c54",
		Name:            "Bog Raiders",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"swampwalk"},
	})
}
