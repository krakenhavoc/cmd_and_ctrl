package effects

// Glittering Lynx — Creature — Cat {W}, 1/1:
//
//	"Prevent all damage that would be dealt to this creature.
//	 {2}: Until end of turn, this creature loses "Prevent all damage
//	 that would be dealt to this creature." Any player may activate this
//	 ability."
//
// The Glittering pair (glittering_helpers.go): a standing self
// prevention, and an any-player row (CR 602.2) that switches that one
// ability off until end of turn, as a layer-6 effect (CR 613.1f, 611.2a).
// Whoever activates pays the {2} out of their own pool (CR 602.1a). The
// prevention is back next turn, and the cat takes damage normally in the
// meantime.
//
// No simplification.
func init() {
	reps, acts := glitteringAbilities("Glittering Lynx", "{2}")
	Register(Spec{
		OracleID:     "890ffe31-642f-46e3-9f09-c744351653b5",
		Name:         "Glittering Lynx",
		Completeness: CompletenessFull,
		Replacements: reps,
		Activated:    acts,
	})
}
