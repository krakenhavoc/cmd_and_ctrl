package effects

// Scare Tactics — Instant {B} (Exodus, #5626):
//
//	"Creatures you control get +1/+0 until end of turn."
//
// Warrior's Charge's body with +1/+0 (creaturesYouControlGetUntilEOT):
// the creatures you control as it resolves are the ones pumped
// (CR 611.2c), so one that enters later this turn is not, and the bonus
// ends in the cleanup step (CR 514.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d7548770-5c60-4c3f-b1d6-554ab5f931f9",
		Name:         "Scare Tactics",
		Completeness: CompletenessFull,
		OnResolve:    creaturesYouControlGetUntilEOT(1, 0, "Scare Tactics — creatures you control get +1/+0"),
	})
}
