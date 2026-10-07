package effects

// Glittering Lion — Creature — Cat {2}{W}, 2/2:
//
//	"Prevent all damage that would be dealt to this creature.
//	 {3}: Until end of turn, this creature loses "Prevent all damage
//	 that would be dealt to this creature." Any player may activate this
//	 ability."
//
// The Glittering pair (glittering_helpers.go): a standing self
// prevention, and an any-player row (CR 602.2) that switches that one
// ability off until end of turn, as a layer-6 effect (CR 613.1f, 611.2a).
// Whoever activates pays the {3} out of their own pool (CR 602.1a). The
// prevention is back next turn, and the cat takes damage normally in the
// meantime.
//
// No simplification.
func init() {
	reps, acts := glitteringAbilities("Glittering Lion", "{3}")
	Register(Spec{
		OracleID:     "549e6de7-56e9-4f5c-8c88-30e446bc53bb",
		Name:         "Glittering Lion",
		Completeness: CompletenessFull,
		Replacements: reps,
		Activated:    acts,
	})
}
