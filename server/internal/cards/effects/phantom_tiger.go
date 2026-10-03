package effects

// Phantom Tiger — Creature — Cat Spirit {2}{G}, 1/0:
//
//	"This creature enters with two +1/+1 counters on it.
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
		OracleID:     "1755b3d4-0e40-40c4-b913-0960d55d411b",
		Name:         "Phantom Tiger",
		Completeness: CompletenessFull,
		Replacements: phantomReplacements("Phantom Tiger", 2),
	})
}
