package effects

// Amphin Pathmage — Creature — Salamander Wizard {3}{U}, 3/2 (Magic 2015,
// #5556):
//
//	"{2}{U}: Target creature can't be blocked this turn."
//
// Elvenking's Harper's ability at {2}{U}: the shared
// targetCantBeBlockedThisTurn grant, with the target re-checked at
// resolution.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "53e4273c-3292-47a0-90d1-8a0eb0eaf381",
		Name:         "Amphin Pathmage",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}{U}: Target creature can't be blocked this turn.",
			Purpose: ForTargets(CantBeBlockedUntilEOT(0, 0, 0)),
			Cost:    ManaCost("{2}{U}"),
			Targets: TargetCreature("target creature"),
			Effect:  targetCantBeBlockedThisTurn("Amphin Pathmage"),
		}},
	})
}
