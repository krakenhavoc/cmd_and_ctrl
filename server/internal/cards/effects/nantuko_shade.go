package effects

// Nantuko Shade — Creature — Insect Shade {B}{B}, 2/1 (Torment, #5612):
//
//	"{B}: This creature gets +1/+1 until end of turn."
//
// Hoar Shade's shade pump (selfPumpRow), which ends at cleanup (CR
// 514.2) and can be activated any number of times.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "40d02b8c-fb0e-4607-855f-49fbcec2e209",
		Name:         "Nantuko Shade",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{selfPumpRow("Nantuko Shade", "{B}", 1, 1)},
	})
}
