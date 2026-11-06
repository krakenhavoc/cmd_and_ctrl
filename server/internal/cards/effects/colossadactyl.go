package effects

// Colossadactyl — Creature — Dinosaur {2}{G}{G}, 4/5 (The Lost Caverns of
// Ixalan, #4470):
//
//	"Reach, trample"
//
// Both are printed keywords the engine reads off the card (CR 702.17a,
// CR 702.19a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "34ff281a-f294-4351-9665-863a5b2e882b",
		Name:            "Colossadactyl",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", "trample"},
	})
}
