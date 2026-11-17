package effects

// Rootwater Hunter — Creature — Merfolk {2}{U}, 1/1 (Tempest, #5628):
//
//	"{T}: This creature deals 1 damage to any target."
//
// Razorfin Hunter's text in mono-blue: the same {T} row, which can't be
// activated while the Hunter is summoning sick, and the shared
// sourceDealsDamageToEachLegalTarget body, so a target gone in response
// takes nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8e42ed82-8b56-46ad-a1ca-e29689a9d030",
		Name:         "Rootwater Hunter",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: This creature deals 1 damage to any target.",
			Cost:    TapCost(),
			Targets: TargetAny(),
			Purpose: ForTargets(DamageToTarget(0, 1)),
			Effect:  sourceDealsDamageToEachLegalTarget(1),
		}},
	})
}
