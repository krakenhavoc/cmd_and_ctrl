package effects

// Granite Shard — Artifact {3} (Mirrodin, #5606):
//
//	"{3}, {T} or {R}, {T}: This artifact deals 1 damage to any target."
//
// Crystal Shard's shape: one ability with a choice of cost, written as
// two rows so the activator picks the cost by picking the row. Both rows
// resolve as Anaba Shaman's ping: 1 damage from the Shard to the target,
// which is checked again as the ability resolves (CR 608.2b).
//
// No simplification.
func init() {
	const effect = "This artifact deals 1 damage to any target."
	Register(Spec{
		OracleID:     "db81c783-a622-41e1-9dd6-47fab1d76967",
		Name:         "Granite Shard",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{3}, {T}: " + effect,
				Purpose: ForTargets(DamageToTarget(0, 1)),
				Cost:    Plus(ManaCost("{3}"), TapCost()),
				Targets: TargetAny(),
				Effect:  DealDamageToTheTarget(1),
			},
			{
				Label:   "{R}, {T}: " + effect,
				Purpose: ForTargets(DamageToTarget(0, 1)),
				Cost:    Plus(ManaCost("{R}"), TapCost()),
				Targets: TargetAny(),
				Effect:  DealDamageToTheTarget(1),
			},
		},
	})
}
