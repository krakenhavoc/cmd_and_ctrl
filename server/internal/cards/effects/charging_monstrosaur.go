package effects

// Charging Monstrosaur — Creature — Dinosaur {4}{R}, 5/5 (Ixalan, #5532):
//
//	"Trample, haste"
//
// Trample and haste are engine keywords carried as printed data: excess
// combat damage goes on to the player (CR 702.19b), and it can attack
// and {T} the turn it arrives (CR 702.10b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "21dd60d5-76c8-469e-8c4e-112e377d2d59",
		Name:            "Charging Monstrosaur",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample", "haste"},
	})
}
