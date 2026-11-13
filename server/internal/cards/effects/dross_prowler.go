package effects

// Dross Prowler — Creature — Zombie {2}{B}, 2/1 (Mirrodin, #5606):
//
//	"Fear (This creature can't be blocked except by artifact creatures
//	 and/or black creatures.)"
//
// Fear is a printed keyword the block legality check reads (CR 702.36b).
// The cost, the type and the 2/1 ride the card from Scryfall.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a3122c6a-20be-4bb3-ab2f-928143378ea6",
		Name:            "Dross Prowler",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"fear"},
	})
}
