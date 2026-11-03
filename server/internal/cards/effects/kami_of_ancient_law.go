package effects

// Kami of Ancient Law — Creature — Spirit {1}{W}, 2/2:
//
//	"Sacrifice this creature: Destroy target enchantment."
//
// One CR 602 ability whose only cost is the sacrifice, paid at announce,
// so the Kami is gone before the destruction resolves; the target is
// re-checked at resolution and an indestructible enchantment survives.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e937e0f5-ca35-4dce-9f05-8f2075731628",
		Name:         "Kami of Ancient Law",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice Kami of Ancient Law: Destroy target enchantment.",
			Cost:    SacrificeThis(),
			Targets: TargetPermanent("target enchantment", Enchantment()),
			Effect:  destroyFirstLegalTarget,
		}},
	})
}
