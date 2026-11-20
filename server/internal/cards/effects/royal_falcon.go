package effects

// Royal Falcon — Creature — Bird {1}{W}, 1/1
// (Starter 1999, #5620):
//
//	"Flying"
//
// Flying is printed data: it can be blocked only by creatures with
// flying or reach (CR 702.9b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a247a7e7-83bb-46f8-9a31-85b0d8ede919",
		Name:            "Royal Falcon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
