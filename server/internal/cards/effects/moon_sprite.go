package effects

// Moon Sprite — Creature — Faerie {1}{G}, 1/1 (Portal, #5630):
//
//	"Flying"
//
// Flying is an engine keyword carried as printed data: it can be blocked
// only by creatures with flying or reach (CR 702.9b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1e41136b-afa7-45ee-8d04-d0c7514b5387",
		Name:            "Moon Sprite",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
