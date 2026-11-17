package effects

// Land Leeches — Creature — Leech {1}{G}{G}, 2/2
// (The Dark, #5638): "First strike". A printed keyword and nothing else.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "228f0afa-146b-4564-a380-71595cdf1ef4",
		Name:            "Land Leeches",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
	})
}
