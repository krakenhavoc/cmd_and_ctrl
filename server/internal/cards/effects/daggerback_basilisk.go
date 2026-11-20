package effects

// Daggerback Basilisk — Creature — Basilisk {2}{G}, 2/2 (Rise of the
// Eldrazi, #5581):
//
//	"Deathtouch (Any amount of damage this deals to a creature is enough to
//	 destroy it.)"
//
// A keyword-only body; deathtouch is the engine's keyword.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ea4d30c7-7c39-46d5-9b71-a7ebc3e219a1",
		Name:            "Daggerback Basilisk",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
	})
}
