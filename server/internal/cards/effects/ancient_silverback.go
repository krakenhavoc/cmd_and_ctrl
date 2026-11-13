package effects

// Ancient Silverback — Creature — Ape {4}{G}{G}, 6/5 (Urza's Destiny,
// #5621):
//
//	"{G}: Regenerate this creature."
//
// The shared regenerateThisRow at {G}: each activation puts one shield
// on the Ape, and the next time it would be destroyed this turn it is
// tapped, removed from combat and has its damage removed instead
// (CR 701.19a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5f12134c-d626-4a5e-8653-700681a491fc",
		Name:         "Ancient Silverback",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{regenerateThisRow("{G}")},
	})
}
