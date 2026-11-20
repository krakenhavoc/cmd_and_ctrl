package effects

// Dragon's Eye Sentry — Creature — Human Monk {W}, 1/3 (Dragons of Tarkir,
// #5551):
//
//	"Defender, first strike"
//
// Defender and first strike are engine keywords carried as printed
// data (CR 702.3, CR 702.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "661b23fc-78cd-44f7-ba53-0cb89ccca41f",
		Name:            "Dragon's Eye Sentry",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender", "first strike"},
	})
}
