package effects

// Seal of Primordium — Enchantment {1}{G} (EDHREC rank 3667):
//
//	"Sacrifice this enchantment: Destroy target artifact or
//	 enchantment."
//
// Seal of Cleansing in green, line for line: Naturalize on a
// permanent, paid for up front and cashed at instant speed later.
// A sacrifice-this activation with no mana and no tap, the target
// picked at announce and re-checked at resolution (CR 608.2b); the
// Seal is sacrificed at announce, so it can never be its own target.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f14dbb39-c9f9-4f64-b22a-38dba28f5b1e",
		Name:         "Seal of Primordium",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice this enchantment: Destroy target artifact or enchantment.",
			Cost:    SacrificeThis(),
			Targets: TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
			Effect:  b35DestroyChosenTarget,
		}},
	})
}
