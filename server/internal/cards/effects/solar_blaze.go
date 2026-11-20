package effects

// Solar Blaze — Sorcery {2}{R}{W} (War of the Spark, #5520):
//
//	"Each creature deals damage to itself equal to its power."
//
// Wave of Reckoning's text and body: every creature deals itself
// damage equal to its power at once.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "424f3d53-47e7-4212-8267-53279ea8900a",
		Name:         "Solar Blaze",
		Completeness: CompletenessFull,
		OnResolve:    eachCreatureDealsDamageToItselfEqualToItsPower,
	})
}
