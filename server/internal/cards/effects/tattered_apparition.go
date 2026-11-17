package effects

// Tattered Apparition — Creature — Shade {3}{B}, 2/2 (Dominaria United, S82):
//
//	"Flying
//	 {1}{B}: This creature gets +1/+1 until end of turn."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b38c0575-37f7-413c-9dad-c8d7e6a41b90",
		Name:            "Tattered Apparition",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated:       []ActivatedAbility{selfPumpRow("Tattered Apparition", "{1}{B}", 1, 1)},
	})
}
