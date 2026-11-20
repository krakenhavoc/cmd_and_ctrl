package effects

// Explosive Apparatus — Artifact {1} (Shadows over Innistrad, #5543):
//
//	"{3}, {T}, Sacrifice this artifact: It deals 2 damage to any target."
//
// Aeolipile's sacrifice-for-two-damage at {3}.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ab8301da-f9f7-4d90-afae-bd7e60834c22",
		Name:         "Explosive Apparatus",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{3}, {T}, Sacrifice this artifact: It deals 2 damage to any target.",
			Purpose: ForTargets(DamageToTarget(0, 2)),
			Cost:    Plus(ManaCost("{3}"), TapCost(), SacrificeThis()),
			Targets: TargetAny(),
			Effect:  DealDamageToTheTarget(2),
		}},
	})
}
