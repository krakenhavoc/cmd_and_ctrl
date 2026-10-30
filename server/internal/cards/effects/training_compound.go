package effects

// Training Compound — Land (Marvel Super Heroes, #4455):
//
//	"{T}: Add {C}.
//	 {T}: Add {R} or {G}. Activate only if this land entered this turn or
//	 if you control a basic land."
//
// Gathering Place's shape: a colourless row and a gated two-colour row
// (mshA3GatedDual), the gate read when the ability is activated.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "99c70f4e-de8a-426d-99aa-17b2f87625ba",
		Name:         "Training Compound",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			colorlessTapRow(),
			mshA3GatedDual("R", "G"),
		},
	})
}
