package effects

// Cemetery Gate — Creature — Wall {2}{B}, 0/5 (Homelands, #5635):
//
//	"Defender
//	 Protection from black"
//
// Both are engine keywords carried as printed data: defender keeps it
// from attacking (CR 702.3b), and protection from black stops black
// spells and abilities from black sources targeting it and prevents
// damage from black sources (CR 702.16b, CR 702.16e).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "443f8dd1-333d-42c4-a286-302a9496209a",
		Name:            "Cemetery Gate",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender", "protection from black"},
	})
}
