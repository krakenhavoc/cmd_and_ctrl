package effects

// Gorilla Chieftain — Creature — Ape {2}{G}{G}, 3/3 (Alliances, #5634):
//
//	"{1}{G}: Regenerate this creature."
//
// A regeneration shield on itself (CR 701.19a), built by
// regenerateThisRow (regenerate_this.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4137ee86-bbf0-49dc-87c5-b48a384baac4",
		Name:         "Gorilla Chieftain",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{regenerateThisRow("{1}{G}")},
	})
}
