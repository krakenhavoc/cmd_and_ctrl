package effects

// Seal of Fire — Enchantment {R} (Nemesis, #5618):
//
//	"Sacrifice this enchantment: It deals 2 damage to any target."
//
// Pyromania's second ability with no mana in the cost: SacrificeThis()
// and the shared sourceDealsDamageToEachLegalTarget body at 2. The
// sacrifice is a cost, so the Seal is gone before anyone can respond,
// and the damage is dealt by it as it last existed (CR 113.7a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "348a345e-4639-41ca-b015-a5d43459eb64",
		Name:         "Seal of Fire",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice this enchantment: It deals 2 damage to any target.",
			Cost:    SacrificeThis(),
			Targets: TargetAny(),
			Purpose: ForTargets(DamageToTarget(0, 2)),
			Effect:  sourceDealsDamageToEachLegalTarget(2),
		}},
	})
}
