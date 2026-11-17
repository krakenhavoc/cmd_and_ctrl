package effects

// Warrior's Charge — Sorcery {2}{W} (Portal, #5630):
//
//	"Creatures you control get +1/+1 until end of turn."
//
// Warrior's Honor's body as a sorcery (creaturesYouControlGetUntilEOT):
// the creatures you control as it resolves (CR 611.2c), pumped until
// cleanup (CR 514.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5cf28c04-76af-4eb8-9969-366dc8e04690",
		Name:         "Warrior's Charge",
		Completeness: CompletenessFull,
		OnResolve:    creaturesYouControlGetUntilEOT(1, 1, "Warrior's Charge — creatures you control get +1/+1"),
	})
}
