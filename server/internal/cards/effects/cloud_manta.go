package effects

// Cloud Manta — Creature — Fish {3}{U}, 3/2 (Battle for Zendikar,
// #5547):
//
//	"Flying"
//
// Flying is an engine keyword carried as printed data: it can be
// blocked only by creatures with flying or reach (CR 702.9b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "59b6bf51-d15d-4032-85ac-29b234b59fc9",
		Name:            "Cloud Manta",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
