package effects

// Shivan Hellkite — {5}{R}{R} Creature — Dragon 5/5:
//
//	"Flying
//	 {1}{R}: This creature deals 1 damage to any target."
//
// No tap in the cost, so it can ping as often as its controller can
// pay, summoning sick or not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "069be6a0-9545-4fdc-ad40-957fe307db7a",
		Name:            "Shivan Hellkite",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:   "{1}{R}: This creature deals 1 damage to any target.",
			Purpose: ForTargets(DamageToTarget(0, 1)),
			Cost:    ManaCost("{1}{R}"),
			Targets: TargetAny(),
			Effect:  DealDamageToTheTarget(1),
		}},
	})
}
