package effects

// Keening Apparition — Creature — Spirit {1}{W}, 2/2 (Return to Ravnica,
// #5566):
//
//	"Sacrifice this creature: Destroy target enchantment."
//
// Ronom Unicorn's ability: the sacrifice is paid at announce, the target is
// re-checked at resolution, and an indestructible enchantment survives.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b9212596-3f65-41c9-9ee8-1f4cec7f0ecf",
		Name:         "Keening Apparition",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice this creature: Destroy target enchantment.",
			Cost:    SacrificeThis(),
			Targets: TargetPermanent("target enchantment", Enchantment()),
			Effect:  destroyFirstLegalTarget,
		}},
	})
}
