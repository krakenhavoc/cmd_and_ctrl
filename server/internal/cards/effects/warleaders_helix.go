package effects

// Warleader's Helix — Instant {2}{R}{W}:
//
//	"Warleader's Helix deals 4 damage to any target and you gain 4
//	 life."
//
// Vampiric Feast's body: 4 damage to the one target, then 4 life. The life
// gain waits for any prompt the damage asks (CR 608.2c, #4155).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b1ea88f5-162b-4f08-bec3-be1f06dec392",
		Name:         "Warleader's Helix",
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		Purpose:      ForTargets(DamageToTarget(0, 4)),
		OnResolve: ResolveSteps(
			Instruction(s82B48DamageToTheFirstTarget(4)),
			GainLife{Amount: 4},
		),
	})
}
