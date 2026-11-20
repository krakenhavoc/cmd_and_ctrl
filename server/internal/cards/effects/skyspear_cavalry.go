package effects

// Skyspear Cavalry — Creature — Human Soldier {3}{W}{W}, 2/2 (Journey into Nyx, #5558):
//
//	"Flying
//	 Double strike (This creature deals both first-strike and regular
//	 combat damage.)"
//
// Printed keywords only, which the engine enforces.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b5868383-a0fb-40fc-b718-f1b5441c5ce3",
		Name:            "Skyspear Cavalry",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "double strike"},
	})
}
