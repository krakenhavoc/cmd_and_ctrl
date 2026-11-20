package effects

// Reckless Reveler — Creature — Satyr {1}{R}, 2/1 (Born of the Gods,
// #5559):
//
//	"{R}, Sacrifice this creature: Destroy target artifact."
//
// Goblin Replica's ability at {R}.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "17d3ceeb-c972-49bc-a530-cf67c49a7e19",
		Name:         "Reckless Reveler",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{R}, Sacrifice this creature: Destroy target artifact.",
			Cost:    Plus(ManaCost("{R}"), SacrificeThis()),
			Targets: TargetPermanent("target artifact", Artifact()),
			Effect:  destroyFirstLegalTarget,
		}},
	})
}
