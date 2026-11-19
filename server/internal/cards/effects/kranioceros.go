package effects

// Kranioceros — Creature — Beast {4}{R}, 5/2 (Conflux, #5587):
//
//	"{1}{W}: This creature gets +0/+3 until end of turn."
//
// The shared thisGetsRow at +0/+3, ending in the cleanup step (CR
// 514.2), any number of times.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c66342bb-bbc1-4aaf-8d99-0af68292081e",
		Name:         "Kranioceros",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{thisGetsRow("{1}{W}", 0, 3, "Kranioceros")},
	})
}
