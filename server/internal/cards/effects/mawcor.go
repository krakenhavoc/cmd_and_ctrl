package effects

// Mawcor — Creature — Beast {3}{U}{U}, 3/3 (Tempest, #5628):
//
//	"Flying
//	 {T}: This creature deals 1 damage to any target."
//
// Flying is a printed keyword (CR 702.9b). The ping is Razorfin
// Hunter's row: {T} in the cost, so it can't be activated while Mawcor
// is summoning sick, and the shared sourceDealsDamageToEachLegalTarget
// body, so a target gone in response takes nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4c0384d6-21cb-4645-b7a4-3af133e8a9a8",
		Name:            "Mawcor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:   "{T}: This creature deals 1 damage to any target.",
			Cost:    TapCost(),
			Targets: TargetAny(),
			Purpose: ForTargets(DamageToTarget(0, 1)),
			Effect:  sourceDealsDamageToEachLegalTarget(1),
		}},
	})
}
