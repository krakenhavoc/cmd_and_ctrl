package effects

// Wind Drake — Creature — Drake {2}{U}, 2/2 (Portal, #5630):
//
//	"Flying"
//
// Flying is an engine keyword carried as printed data: it can be blocked
// only by creatures with flying or reach (CR 702.9b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d6ffdaf0-ac08-4de9-bbce-2eab2f86bcca",
		Name:            "Wind Drake",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
