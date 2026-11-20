package effects

// Highspire Mantis — Creature — Insect {2}{R}{W}, 3/3 (Khans of Tarkir,
// #5554):
//
//	"Flying, trample"
//
// Flying (CR 702.9b) and trample (CR 702.19b) are engine keywords
// carried as printed data.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6adff11d-5126-4a4b-b182-e2700d406995",
		Name:            "Highspire Mantis",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
	})
}
