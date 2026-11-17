package effects

// Desert Drake — Creature — Drake {3}{R}, 2/2 (Portal, #5630):
//
//	"Flying"
//
// Flying is an engine keyword carried as printed data: it can be blocked
// only by creatures with flying or reach (CR 702.9b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ce8f4eb4-08b8-404b-9147-1e28c1b14a65",
		Name:            "Desert Drake",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
