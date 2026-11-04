package effects

// Ronom Unicorn — Creature — Unicorn {1}{W}, 2/2:
//
//	"Sacrifice this creature: Destroy target enchantment."
//
// Kami of Ancient Law's ability under another name: the sacrifice is
// paid at announce, so the Unicorn is gone before the destruction
// resolves; the target is re-checked at resolution and an
// indestructible enchantment survives.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3c6913cc-c543-45f8-ad3b-ccd282efa6f7",
		Name:         "Ronom Unicorn",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice Ronom Unicorn: Destroy target enchantment.",
			Cost:    SacrificeThis(),
			Targets: TargetPermanent("target enchantment", Enchantment()),
			Effect:  destroyFirstLegalTarget,
		}},
	})
}
