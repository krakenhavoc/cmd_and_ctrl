package effects

// Bog Wraith — Creature — Wraith {3}{B}, 3/3 (Portal, #5630):
//
//	"Swampwalk (This creature can't be blocked as long as defending player controls a Swamp.)"
//
// Swampwalk is an engine keyword carried as printed data: it can't be
// blocked while the defending player controls a Swamp (CR 702.14c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "508248d1-09a4-4e41-a4c9-286618e5061e",
		Name:            "Bog Wraith",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"swampwalk"},
	})
}
