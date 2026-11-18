package effects

// Conclave Tribunal — Enchantment {3}{W}:
//
//	"Convoke (Your creatures can help cast this spell. Each creature you
//	 tap while casting this spell pays for {1} or one mana of that
//	 creature's color.)
//	 When this enchantment enters, exile target nonland permanent an
//	 opponent controls until this enchantment leaves the battlefield."
//
// Banishing Light's text behind convoke (CR 702.51): the exile records its
// own return (CR 610.3), so the permanent comes back as a new object under
// its owner's control the moment this leaves, and nothing is exiled if the
// enchantment left first (CR 610.3b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e938ee4c-d5df-4d93-bd61-9e518fb1dc30",
		Name:         "Conclave Tribunal",
		Completeness: CompletenessFull,
		TapCost:      Convoke(),
		Triggered:    exileUntilThisLeavesTriggers("Conclave Tribunal"),
	})
}
