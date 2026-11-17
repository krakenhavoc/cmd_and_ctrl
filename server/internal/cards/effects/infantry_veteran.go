package effects

// Infantry Veteran — Creature — Human Soldier {W}, 1/1 (Visions, #5631):
//
//	"{T}: Target attacking creature gets +1/+1 until end of turn."
//
// The {T} keeps it off the turn it arrives (CR 602.5a). Only an
// attacking creature can be the target; the boost is a layer 7c change
// (CR 613.4c) that ends at cleanup (CR 514.2), and a target no longer
// attacking or gone gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "42798e2b-9a5c-4f93-8f77-b7bb7a916d07",
		Name:         "Infantry Veteran",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: Target attacking creature gets +1/+1 until end of turn.",
			Purpose: ForTargets(BenefitUntilEOT(0, 1, 1)),
			Cost:    TapCost(),
			Targets: TargetCreature("target attacking creature", AttackingCreature()),
			Effect:  pumpAndGrantFirstTarget(1, 1, "Infantry Veteran — +1/+1 until end of turn"),
		}},
	})
}
