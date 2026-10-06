package effects

// Benalish Missionary — Creature — Human Cleric {W}, 1/1:
//
//	"{1}{W}, {T}: Prevent all combat damage that would be dealt by target
//	 blocked creature this turn."
//
// "Blocked creature" is CR 509.1h's blocked record, read through
// game.BlockedAttackerForEffect (#2026): an attacker that was blocked
// stays blocked after its blockers are gone, and one "blocked" by an
// effect counts (the ruling). The shield's source is the target, pinned
// as the ability resolves (CR 400.7), as Kor Haven's is.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "84fdcfd3-2b22-4570-af34-7e3f55f97466",
		Name:         "Benalish Missionary",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{shieldAgainstTargetsRow(
			"{1}{W}, {T}: Prevent all combat damage that would be dealt by target blocked creature this turn.",
			Plus(ManaCost("{1}{W}"), TapCost()),
			TargetCreature("target blocked creature", BlockedCreature()), true)},
	})
}
