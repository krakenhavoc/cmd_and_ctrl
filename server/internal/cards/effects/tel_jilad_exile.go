package effects

// Tel-Jilad Exile — Creature — Troll Warrior {3}{G}, 2/3 (Mirrodin, #5606):
//
//	"{1}{G}: Regenerate this creature."
//
// The shared regenerate row (regenerate_this.go): a shield on the
// ability's source, used the next time it would be destroyed this
// turn (CR 701.19a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "58987ca8-ee9b-412e-b08f-440588234e00",
		Name:         "Tel-Jilad Exile",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{regenerateThisRow("{1}{G}")},
	})
}
