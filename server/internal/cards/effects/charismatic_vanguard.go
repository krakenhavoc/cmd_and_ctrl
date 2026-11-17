package effects

// Charismatic Vanguard — Creature — Dwarf Soldier {2}{W}, 3/2 (Dominaria United, S82):
//
//	"{4}{W}: Creatures you control get +1/+1 until end of turn."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f1861fac-e0e5-4b96-b6e4-6b6573bd02d1",
		Name:         "Charismatic Vanguard",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{s82DmuB01TeamPumpRow("Charismatic Vanguard", "{4}{W}")},
	})
}
