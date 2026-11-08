package effects

// Swiftblade Vindicator — Creature — Human Soldier {R}{W}, 1/1
// (Foundations, #4465):
//
//	"Double strike
//	 Vigilance
//	 Trample"
//
// Three printed keywords the engine enforces from PrintedKeywords: it
// deals first-strike and regular combat damage, attacking doesn't tap
// it, and its excess combat damage tramples over.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "44240b1e-c104-4eb4-8ab9-0d56607ba241",
		Name:            "Swiftblade Vindicator",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"double strike", "vigilance", "trample"},
	})
}
