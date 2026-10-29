package effects

// Elvenking's Harper — Creature — Elf Bard {1}{U}, 2/2 (The Hobbit):
//
//	"{4}{U}: Target creature can't be blocked this turn."
//
// One activated ability with no timing gate: any creature, any time you
// have priority, so it is also an instant-speed answer to a blocker that
// would otherwise stop an attacker already declared. The grant is the
// same RestrictUntilEOT that Zhuge Jin and Access Tunnel use
// (targetCantBeBlockedThisTurn), and the target is re-checked at
// resolution like any target.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7d070a95-49ea-4aea-b512-dbb0006a00ee",
		Name:         "Elvenking's Harper",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{4}{U}: Target creature can't be blocked this turn.",
			Purpose: ForTargets(CantBeBlockedUntilEOT(0, 0, 0)),
			Cost:    ManaCost("{4}{U}"),
			Targets: TargetCreature("target creature"),
			Effect:  targetCantBeBlockedThisTurn("Elvenking's Harper"),
		}},
	})
}
