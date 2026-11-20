package effects

// Broodhunter Wurm — Creature — Wurm {3}{G}, 4/3 (Battle for Zendikar,
// #5547):
//
//	(no rules text)
//
// A vanilla body: no rules text, nothing to register beyond the card
// itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "02d50c11-51e3-4b3f-bf3d-3b1775d514f6",
		Name:         "Broodhunter Wurm",
		Completeness: CompletenessFull,
	})
}
