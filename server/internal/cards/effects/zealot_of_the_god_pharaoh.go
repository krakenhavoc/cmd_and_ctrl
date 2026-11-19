package effects

// Zealot of the God-Pharaoh — Creature — Minotaur Archer {3}{R}, 4/3:
//
//	"{4}{R}: This creature deals 2 damage to target opponent or
//	 planeswalker."
//
// The target is an opponent (never you or a teammate) or a planeswalker.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "2911d52f-7faf-455a-acdf-77269c9ee8ce",
		Name:         "Zealot of the God-Pharaoh",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{4}{R}: This creature deals 2 damage to target opponent or planeswalker.",
			Purpose: ForTargets(DamageToTarget(0, 2)),
			Cost:    ManaCost("{4}{R}"),
			Targets: targetOpponentOrPlaneswalker(),
			Effect:  DealDamageToTheTarget(2),
		}},
	})
}
