package effects

// Tower of Calamities — Artifact {4} (Scars of Mirrodin, #5577):
//
//	"{8}, {T}: This artifact deals 12 damage to target creature."
//
// Aladdin's Ring's shape at 12 to a creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "27618701-e8c9-4117-94f3-3836cf34990d",
		Name:         "Tower of Calamities",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{8}, {T}: This artifact deals 12 damage to target creature.",
			Purpose: ForTargets(DamageToTarget(0, 12)),
			Cost:    Plus(ManaCost("{8}"), TapCost()),
			Targets: TargetCreature("target creature"),
			Effect:  DealDamageToTheTarget(12),
		}},
	})
}
