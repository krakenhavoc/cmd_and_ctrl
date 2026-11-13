package effects

// Benalish Knight — Creature — Human Knight {2}{W}, 2/2 (Weatherlight, #5629):
//
//	"Flash (You may cast this spell any time you could cast an instant.)
//	 First strike (This creature deals combat damage before creatures
//	 without first strike.)"
//
// Both are printed keywords: flash gates the cast from hand (CR 702.8a)
// and first strike splits the combat damage step (CR 702.7b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "32b401e9-163f-4917-a728-fc63b25ef602",
		Name:            "Benalish Knight",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "first strike"},
	})
}
