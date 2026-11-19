package effects

// Void Helix — Instant {4}{B}{B} (Marvel Super Heroes Commander,
// #4455): "Void Helix deals 5 damage to any target and you gain 5
// life." Lightning Helix's shape at five: the life gain waits for a
// prompt the damage asks (CR 608.2c), so it is its own step after the
// damage.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "82e52a09-ce6f-417d-8460-8806df63f0bf",
		Name:         "Void Helix",
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		Purpose:      ForTargets(DamageToTarget(0, 5)),
		OnResolve: ResolveSteps(
			Instruction(s82B48DamageToTheFirstTarget(5)),
			GainLife{Amount: 5},
		),
	})
}
