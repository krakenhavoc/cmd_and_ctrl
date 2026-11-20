package effects

// Vampiric Feast — Sorcery {5}{B}{B} (Portal, #5630):
//
//	"Vampiric Feast deals 4 damage to any target and you gain 4 life."
//
// Void Helix's body with 4: the damage, then the life gain.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1980ca2e-a415-4de1-ac30-7055507e82a2",
		Name:         "Vampiric Feast",
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		Purpose:      ForTargets(DamageToTarget(0, 4)),
		OnResolve: ResolveSteps(
			Instruction(s82B48DamageToTheFirstTarget(4)),
			GainLife{Amount: 4},
		),
	})
}
