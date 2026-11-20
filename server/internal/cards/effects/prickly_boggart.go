package effects

// Prickly Boggart — Creature — Goblin Rogue {B}, 1/1 (Morningtide, #5592):
//
//	"Fear (This creature can't be blocked except by artifact creatures
//	 and/or black creatures.)"
//
// Razortooth Rats' printed fear (CR 702.36b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "fa110e05-6759-4eeb-96eb-9b7a8e4a94c3",
		Name:            "Prickly Boggart",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"fear"},
	})
}
