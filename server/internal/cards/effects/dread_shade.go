package effects

// Dread Shade — Creature — Shade {B}{B}{B}, 3/3 (Dominaria, #5530):
//
//	"{B}: This creature gets +1/+1 until end of turn."
//
// Nantuko Shade's pump, the shared selfPumpRow.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9b90c7d1-38c6-4210-bc76-68c5fcb91202",
		Name:         "Dread Shade",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{selfPumpRow("Dread Shade", "{B}", 1, 1)},
	})
}
