package effects

// Felidar Cub — Creature — Cat Beast {1}{W}, 2/2 (Foundations, #4465):
//
//	"Sacrifice this creature: Destroy target enchantment."
//
// Ronom Unicorn's ability: the sacrifice is paid at announce, so the Cub
// is gone before the destruction resolves; the target is checked again as
// it resolves (CR 608.2b), and an indestructible enchantment survives.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cae60cf8-cd64-4595-a4dd-946694cf2bb1",
		Name:         "Felidar Cub",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice this creature: Destroy target enchantment.",
			Cost:    SacrificeThis(),
			Targets: TargetPermanent("target enchantment", Enchantment()),
			Effect:  destroyFirstLegalTarget,
		}},
	})
}
