package effects

// Almighty Brushwagg — Creature — Brushwagg {G}, 1/1 (Ikoria: Lair of Behemoths, #5512):
//
//	"Trample
//	 {3}{G}: This creature gets +3/+3 until end of turn."
//
// The firebreathing row (selfPumpRow): no tap, so any number of times.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d8a3a89d-9d27-40ae-a1aa-d55cc6a79b0b",
		Name:            "Almighty Brushwagg",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Activated:       []ActivatedAbility{selfPumpRow("Almighty Brushwagg", "{3}{G}", 3, 3)},
	})
}
