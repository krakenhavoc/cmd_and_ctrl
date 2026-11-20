package effects

// Molten Ravager — Creature — Elemental {2}{R}, 0/4 (Zendikar, #5583):
//
//	"{R}: This creature gets +1/+0 until end of turn."
//
// The shared firebreathing row (self_pump_row.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "697c851d-8475-479e-8783-cebdbe9a892d",
		Name:         "Molten Ravager",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{selfPumpRow("Molten Ravager", "{R}", 1, 0)},
	})
}
