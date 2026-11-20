package effects

// Raiding Nightstalker — Creature — Nightstalker {2}{B}, 2/2 (Portal Second Age,
// #5625):
//
//	"Swampwalk (This creature can't be blocked as long as defending player controls a Swamp.)"
//
// Swampwalk is an engine keyword carried as printed data: it can't be
// blocked while the defending player controls a Swamp (CR 702.14c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5c4cb33d-2a79-4934-839f-d34e9ca284f8",
		Name:            "Raiding Nightstalker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"swampwalk"},
	})
}
