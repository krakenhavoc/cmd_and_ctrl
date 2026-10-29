package effects

// Web Up — Enchantment {2}{W} (Marvel Super Heroes, #4455):
//
//	"When this enchantment enters, exile target nonland permanent an
//	 opponent controls until this enchantment leaves the battlefield."
//
// Banishing Light under a Marvel name (Oblivion Ring's shape, CR 610.3):
// the rows are exileUntilThisLeavesTriggers, so the exiled permanent
// comes back as a new object under its owner's control the moment this
// leaves, and nothing is exiled if it leaves before the trigger resolves
// (CR 610.3b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9b9b086c-d201-4574-a1e9-0d6f8de53d7d",
		Name:         "Web Up",
		Completeness: CompletenessFull,
		Triggered:    exileUntilThisLeavesTriggers("Web Up"),
	})
}
