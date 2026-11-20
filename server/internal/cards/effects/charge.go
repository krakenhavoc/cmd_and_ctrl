package effects

// Charge — Instant {W} (Dominaria, #5530):
//
//	"Creatures you control get +1/+1 until end of turn."
//
// Warrior's Honor's team pump: the creatures you control as it resolves
// get +1/+1 until end of turn (CR 611.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "36b5022d-fa62-46ea-97ab-23c0f60f62a5",
		Name:         "Charge",
		Completeness: CompletenessFull,
		OnResolve:    creaturesYouControlGetUntilEOT(1, 1, "Charge — creatures you control get +1/+1"),
	})
}
