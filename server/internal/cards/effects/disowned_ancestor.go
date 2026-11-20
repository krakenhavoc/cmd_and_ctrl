package effects

// Disowned Ancestor — Creature — Spirit Warrior {B}, 0/4 (Khans of
// Tarkir, #5554):
//
//	"Outlast {1}{B} ({1}{B}, {T}: Put a +1/+1 counter on this creature.
//	 Outlast only as a sorcery.)"
//
// Outlast is the shared sorcery-speed tap-and-pay counter row
// (rprB04Outlast, CR 702.107a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "192603a4-720e-4c53-bee5-4f2826c1ed7f",
		Name:         "Disowned Ancestor",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{rprB04Outlast("{1}{B}")},
	})
}
