package effects

// Warrior's Honor — Instant {2}{W} (Visions, #5631):
//
//	"Creatures you control get +1/+1 until end of turn."
//
// The creatures you control as it resolves (CR 611.2c), so one that
// enters later this turn is not pumped. A layer 7c change (CR 613.4c)
// that ends at cleanup (CR 514.2).
//
// The shared body is creaturesYouControlGetUntilEOT (#5630).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c3c9da49-ef4f-48c7-99c8-0f226b485285",
		Name:         "Warrior's Honor",
		Completeness: CompletenessFull,
		OnResolve:    creaturesYouControlGetUntilEOT(1, 1, "Warrior's Honor — creatures you control get +1/+1"),
	})
}
