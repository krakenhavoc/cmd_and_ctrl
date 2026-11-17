package effects

// Storm Spirit — Creature — Elemental Spirit {3}{G}{W}{U}, 3/3 (Ice Age,
// #5636):
//
//	"Flying
//	 {T}: This creature deals 2 damage to target creature."
//
// Flying is a printed keyword. The tap ability is any creature, the
// shared DealDamageToTheTarget body, dealt by the Spirit; a target that
// left in response is left alone (CR 608.2b). The tap means summoning
// sickness applies (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6a413d09-dd67-4c3e-b81b-7bfe8b864dc9",
		Name:            "Storm Spirit",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:   "{T}: This creature deals 2 damage to target creature.",
			Purpose: ForTargets(DamageToTarget(0, 2)),
			Cost:    TapCost(),
			Targets: TargetCreature("target creature"),
			Effect:  DealDamageToTheTarget(2),
		}},
	})
}
