package effects

// Stormplain Detainment — Enchantment {2}{W} (Tarkir: Dragonstorm,
// #4463):
//
//	"When this enchantment enters, exile target nonland permanent an
//	 opponent controls until this enchantment leaves the battlefield."
//
// Banishing Light's text, so Banishing Light's triggers
// (exileUntilThisLeavesTriggers, exile_until.go). If the Detainment has
// already left the battlefield when the trigger resolves, nothing is
// exiled (CR 610.3b); when it leaves later, the card returns under its
// owner's control (CR 610.3c) as a new object.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4ad9c550-5605-4ffe-8e27-e9f2d9ada000",
		Name:         "Stormplain Detainment",
		Completeness: CompletenessFull,
		Triggered:    exileUntilThisLeavesTriggers("Stormplain Detainment"),
	})
}
