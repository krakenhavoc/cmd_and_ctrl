package effects

// Arc-Slogger — Creature — Beast {3}{R}{R}, 4/5:
//
//	"{R}, Exile the top ten cards of your library: This creature deals 2
//	 damage to any target."
//
// ADR 0109 §7 (#1902): "Exile the top N cards of your library" is a
// cost with nothing to choose. A library of fewer than N cards can't pay
// it (CR 118.3), and it is paid after every other cost (CR 601.2h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a28fc506-a98c-4b4d-a0f4-7d971489ccbb",
		Name:         "Arc-Slogger",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{R}, Exile the top ten cards of your library: This creature deals 2 damage to any target.",
			Cost:    Plus(ManaCost("{R}"), ExileTopOfLibrary(10)),
			Targets: TargetAny(),
			Purpose: ForTargets(DamageToTarget(0, 2)),
			Effect:  sourceDealsDamageToEachLegalTarget(2),
		}},
	})
}
