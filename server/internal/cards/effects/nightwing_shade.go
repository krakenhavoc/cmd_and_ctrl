package effects

// Nightwing Shade — Creature — Shade {4}{B}, 2/2 (Magic 2011, #5579):
//
//	"Flying
//	 {1}{B}: This creature gets +1/+1 until end of turn."
//
// Flying is the printed keyword. The pump is Thornling's shared
// thisGetsRow, a layer 7c change that ends at cleanup and stacks with
// each activation.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e679ef01-8d12-4246-850a-efed7f4316b2",
		Name:            "Nightwing Shade",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated:       []ActivatedAbility{thisGetsRow("{1}{B}", 1, 1, "Nightwing Shade")},
	})
}
