package effects

// Solidarity — Instant {3}{W} (Urza's Destiny, #5621):
//
//	"Creatures you control get +0/+5 until end of turn."
//
// Steadfastness at +0/+5, through the shared
// creaturesYouControlGetUntilEOT body: the creatures you control as it
// resolves are the ones boosted (CR 611.2c), and the boost ends in the
// cleanup step (CR 514.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7b094a2b-2b69-4a9b-9829-c1aabc53f75e",
		Name:         "Solidarity",
		Completeness: CompletenessFull,
		OnResolve:    creaturesYouControlGetUntilEOT(0, 5, "Solidarity — creatures you control get +0/+5"),
	})
}
