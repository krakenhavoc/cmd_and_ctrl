package effects

// Scavenger Folk — Creature — Human {G}, 1/1 (The Dark, #5638):
//
//	"{G}, {T}, Sacrifice this creature: Destroy target artifact."
//
// A mana, tap and sacrifice cost with Shatter's target clause.
// Artifact() reads the current types, so an animated artifact is a
// legal target.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "74f361f8-a0d8-4824-9380-61469d6a19cf",
		Name:         "Scavenger Folk",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{G}, {T}, Sacrifice this creature: Destroy target artifact.",
			Cost:    Plus(ManaCost("{G}"), TapCost(), SacrificeThis()),
			Targets: TargetPermanent("target artifact", Artifact()),
			Effect:  destroyFirstLegalTarget,
		}},
	})
}
