package effects

// Flame Javelin — Instant {2/R}{2/R}{2/R} (Shadowmoor, #5591):
//
//	"({2/R} can be paid with any two mana or with {R}. This card's mana
//	 value is 6.)
//	 Flame Javelin deals 4 damage to any target."
//
// Puncture Blast's damage row without wither: 4 to any target, the target
// re-checked at resolution (CR 608.2b). Each {2/R} is paid with {R} or two
// mana (CR 107.4e).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "58363090-bf89-4174-bbb2-636afc1c7a06",
		Name:         "Flame Javelin",
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		Purpose:      ForTargets(DamageToTarget(0, 4)),
		OnResolve:    damageToFirstTarget(4),
	})
}
