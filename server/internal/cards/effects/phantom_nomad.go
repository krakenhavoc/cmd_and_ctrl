package effects

// Phantom Nomad — Creature — Spirit Nomad {1}{W}, 0/0:
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
		OracleID:     "bf65a7a7-a590-41eb-9844-184e5d63e32a",
		Name:         "Phantom Nomad",
		Completeness: CompletenessFull,
		Replacements: phantomReplacements("Phantom Nomad", 2),
	})
}
