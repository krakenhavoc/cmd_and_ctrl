package effects

// Pyrite Spellbomb — Artifact {1} (Mirrodin, #5606):
//
//	"{R}, Sacrifice this artifact: It deals 2 damage to any target.
//	 {1}, Sacrifice this artifact: Draw a card."
//
// The first row is Aeolipile's: 2 damage from the Spellbomb, read as
// it last existed (CR 608.2h), to a target checked again as the
// ability resolves (CR 608.2b).
// The second row is Aether Spellbomb's: {1}, sacrifice it, draw a card
// (mrdB5SpellbombDraw). The sacrifice is part of the cost, so the
// Spellbomb is gone before either row resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2c10cae2-951a-4f4f-94e4-8713b58d07dd",
		Name:         "Pyrite Spellbomb",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{R}, Sacrifice this artifact: It deals 2 damage to any target.",
				Purpose: ForTargets(DamageToTarget(0, 2)),
				Cost:    Plus(ManaCost("{R}"), SacrificeThis()),
				Targets: TargetAny(),
				Effect:  DealDamageToTheTarget(2),
			},
			mrdB5SpellbombDraw(),
		},
	})
}
