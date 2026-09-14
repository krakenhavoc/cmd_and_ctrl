package effects

// Thrashing Brontodon — Creature — Dinosaur {1}{G}{G}, 3/4 (EDHREC
// rank 2799):
//
//	"{1}, Sacrifice this creature: Destroy target artifact or
//	 enchantment."
//
// The green deck's body that is also its Naturalize. One activated
// ability, sacrificing the Brontodon as its cost — so it dies with
// the ability on the stack and a Blood Artist drains before the
// artifact is destroyed — through the single-target verb that
// honours indestructible. A target that left in response makes the
// ability do nothing, and the Brontodon is gone either way, as
// printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "60bc63dc-ac9f-4a2f-aef5-c90d0aa31553",
		Name:         "Thrashing Brontodon",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}, Sacrifice this creature: Destroy target artifact or enchantment.",
			Cost:    Plus(ManaCost("{1}"), SacrificeThis()),
			Targets: TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
			Effect:  b17DestroyFirstLegalTarget,
		}},
	})
}
