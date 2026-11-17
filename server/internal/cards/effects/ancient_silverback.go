package effects

// Ancient Silverback — Creature — Ape {4}{G}{G}, 6/5 (Urza's Destiny,
// #5621):
//
//	"{G}: Regenerate this creature."
//
// Gorilla Chieftain's row (regenerateThisRow) on a 6/5: each activation
// sets up one shield for this turn (CR 701.19a), used up the next time
// the Ape would be destroyed. No {T} in the cost, so it can be activated
// as often as you can pay, summoning sick or not.
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
