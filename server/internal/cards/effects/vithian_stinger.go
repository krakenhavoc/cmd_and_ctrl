package effects

// Vithian Stinger — Creature — Human Shaman {2}{R}, 0/1 (Shards of Alara,
// #5588):
//
//	"{T}: This creature deals 1 damage to any target.
//	 Unearth {1}{R}"
//
// Thornwind Faeries' ping with the shared Unearth row (CR 702.84a). The
// ping has {T} in its cost, so it waits out summoning sickness unless the
// Stinger has haste, which an unearthed one does. A target gone in
// response takes nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1c329131-bef2-41c6-823a-136c110ff0ad",
		Name:         "Vithian Stinger",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{T}: This creature deals 1 damage to any target.",
				Cost:    TapCost(),
				Targets: TargetAny(),
				Purpose: ForTargets(DamageToTarget(0, 1)),
				Effect:  sourceDealsDamageToEachLegalTarget(1),
			},
			Unearth("{1}{R}"),
		},
	})
}
