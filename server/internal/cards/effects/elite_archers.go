package effects

// Elite Archers — Creature — Human Soldier Archer {5}{W}, 3/3
// (Urza's Saga):
//
//	"{T}: This creature deals 3 damage to target attacking or blocking
//	 creature."
//
// Heavy Ballista's row for 3.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2230ec18-4dcf-435e-b495-1905cb94e8b8",
		Name:         "Elite Archers",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: This creature deals 3 damage to target attacking or blocking creature.",
			Cost:    TapCost(),
			Targets: TargetCreature("target attacking or blocking creature", AttackingOrBlocking()),
			Purpose: ForTargets(DamageToTarget(0, 3)),
			Effect:  sourceDealsDamageToEachLegalTarget(3),
		}},
	})
}
