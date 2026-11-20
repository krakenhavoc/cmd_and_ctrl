package effects

// Sylvok Replica — Artifact Creature — Shaman {3}, 1/3 (Scars of Mirrodin, #5577):
//
//	"{G}, Sacrifice this creature: Destroy target artifact or enchantment."
//
// Thrashing Brontodon's ability at {G}. The Replica is sacrificed as a
// cost, so it can't be its own target.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "97818ed4-3590-4a3b-92ce-716f9cfdc78e",
		Name:         "Sylvok Replica",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{G}, Sacrifice this creature: Destroy target artifact or enchantment.",
			Cost:    Plus(ManaCost("{G}"), SacrificeThis()),
			Targets: TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
			Effect:  destroyFirstLegalTarget,
		}},
	})
}
