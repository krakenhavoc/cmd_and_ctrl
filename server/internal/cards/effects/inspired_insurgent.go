package effects

// Inspired Insurgent — Creature — Human Peasant Ally {1}{W}, 2/2
// (Avatar: The Last Airbender Eternal, #5075):
//
//	"{1}, Sacrifice this creature: Destroy target artifact or enchantment."
//
// Cathar Commando's ability (destroyChosenPermanent): the sacrifice is a
// cost, so the destroy still happens after the Insurgent is gone.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5863c67b-a47c-4466-8bd1-210ab26e0d3f",
		Name:         "Inspired Insurgent",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}, Sacrifice this creature: Destroy target artifact or enchantment.",
			Cost:    Plus(ManaCost("{1}"), SacrificeThis()),
			Targets: TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
			Effect:  destroyChosenPermanent,
		}},
	})
}
