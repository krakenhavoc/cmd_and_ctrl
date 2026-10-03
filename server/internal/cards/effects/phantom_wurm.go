package effects

// Phantom Wurm — Creature — Wurm Spirit {4}{G}{G}, 2/0:
//
//	"This creature enters with four +1/+1 counters on it.
//	 If damage would be dealt to this creature, prevent that damage. Remove a +1/+1 counter from this creature."
//
// ADR 0108 §8 (#1906): one of the Phantoms (phantoms.go). Damage dealt to
// it by any number of sources at once is all prevented and costs one
// counter; damage that can't be prevented is dealt and still costs one
// (CR 615.12); with no counter left the damage is still prevented (the
// rulings).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "a866fb67-6614-45e6-928e-b4b59d95335f",
		Name:         "Phantom Wurm",
		Completeness: CompletenessFull,
		Replacements: phantomReplacements("Phantom Wurm", 4),
	})
}
