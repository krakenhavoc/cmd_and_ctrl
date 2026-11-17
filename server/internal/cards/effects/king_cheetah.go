package effects

// King Cheetah — Creature — Cat {3}{G}, 3/2 (Multiverse Gift Box, #5632):
//
//	"Flash"
//
// The printed keyword is the whole card: flash lets it be cast any time
// its controller could cast an instant (CR 702.8a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f9ca35e1-24c5-46f4-a666-0dfe2c920e8b",
		Name:            "King Cheetah",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
	})
}
