package effects

// Caustic Caterpillar — Creature — Insect {G}, 1/1 (EDHREC rank
// 3436):
//
//	"{1}{G}, Sacrifice this creature: Destroy target artifact or
//	 enchantment."
//
// A Naturalize on legs. One CR 602 ability with a mana-plus-self-
// sacrifice cost, paid at announce so the Caterpillar's own
// dies-triggers resolve above the destruction; the target is
// re-checked at resolution and an indestructible pick survives, as
// printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "45d35128-76e6-43f9-8d23-41f7506c3a71",
		Name:         "Caustic Caterpillar",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}{G}, Sacrifice Caustic Caterpillar: Destroy target artifact or enchantment.",
			Cost:    Plus(ManaCost("{1}{G}"), SacrificeThis()),
			Targets: TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
			Effect:  destroyFirstLegalTarget,
		}},
	})
}
