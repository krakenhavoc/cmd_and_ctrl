package effects

// Seacoast Drake — Creature — Drake {1}{U}, 1/3 (Magic 2014, #5563):
//
//	"Flying"
//
// A 1/3 flier. Flying is a canonical keyword the block check reads.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2f4b5831-57a2-4778-90c1-0b8ec49afd46",
		Name:            "Seacoast Drake",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
