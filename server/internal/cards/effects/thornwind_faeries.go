package effects

// Thornwind Faeries — Creature — Faerie {1}{U}{U}, 1/1 (Urza's Legacy,
// #5623):
//
//	"Flying
//	 {T}: This creature deals 1 damage to any target."
//
// Mawcor's row on a 1/1 flier: flying is a printed keyword (CR 702.9b), and
// the ping has {T} in its cost, so it waits out summoning sickness
// (CR 302.6). A target gone in response takes nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "941a0e8d-c407-476b-9d0e-cffc1f8c6003",
		Name:            "Thornwind Faeries",
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
