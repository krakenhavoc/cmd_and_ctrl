package effects

// Righteous Charge — Sorcery {1}{W}{W} (Portal Second Age, #5625):
//
//	"Creatures you control get +2/+2 until end of turn."
//
// Warrior's Charge's body at +2/+2 (creaturesYouControlGetUntilEOT): the
// creatures you control as it resolves are the ones pumped (CR 611.2c),
// and the bonus ends in the cleanup step (CR 514.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "252d9183-16eb-4345-8d15-56218b33abdd",
		Name:         "Righteous Charge",
		Completeness: CompletenessFull,
		OnResolve:    creaturesYouControlGetUntilEOT(2, 2, "Righteous Charge — creatures you control get +2/+2"),
	})
}
