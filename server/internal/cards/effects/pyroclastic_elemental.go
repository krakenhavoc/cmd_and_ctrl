package effects

// Pyroclastic Elemental — Creature — Elemental {3}{R}{R}, 5/4 (Core Set 2020, #5518):
//
//	"{1}{R}{R}: This creature deals 1 damage to target player."
//
// A player-only ping with no {T}, so it can be activated any number of
// times.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1f37e3a6-ce3a-472f-ac45-0d9b23a31a11",
		Name:         "Pyroclastic Elemental",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}{R}{R}: This creature deals 1 damage to target player.",
			Purpose: ForTargets(DamageToTarget(0, 1)),
			Cost:    ManaCost("{1}{R}{R}"),
			Targets: TargetPlayer("target player"),
			Effect:  DealDamageToTheTarget(1),
		}},
	})
}
