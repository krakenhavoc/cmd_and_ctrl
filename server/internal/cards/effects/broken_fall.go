package effects

// Broken Fall — {2}{G} Enchantment:
//
//	"Return this enchantment to its owner's hand: Regenerate target
//	 creature."
//
// #2028: the return is the cost (ReturnThis), paid at announce
// (CR 602.2b). Molting Skin prints the same sentence; both use
// returnThisRegenerateRow.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "79791c7f-dd33-406a-8c2e-6b722abd5161",
		Name:         "Broken Fall",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{returnThisRegenerateRow()},
	})
}
