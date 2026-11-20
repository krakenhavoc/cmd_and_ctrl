package effects

// Tidehollow Strix — Artifact Creature — Bird {U}{B}, 2/1 (Shards of
// Alara, #5588):
//
//	"Flying
//	 Deathtouch"
//
// Two printed keywords: flying (CR 702.9b) and deathtouch (CR 702.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "19a39cc5-2134-4dee-932c-81389806c57b",
		Name:            "Tidehollow Strix",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "deathtouch"},
	})
}
