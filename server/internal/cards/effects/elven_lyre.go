package effects

// Elven Lyre — Artifact {2} (Fallen Empires, #5637):
//
//	"{1}, {T}, Sacrifice this artifact: Target creature gets +2/+2 until
//	 end of turn."
//
// The Lyre is sacrificed as part of the cost (CR 602.2), so it is gone
// before the pump resolves; a target that left in response gets
// nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7601378d-42cb-4351-8316-8f78a3b49a85",
		Name:         "Elven Lyre",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}, {T}, Sacrifice this artifact: Target creature gets +2/+2 until end of turn.",
			Purpose: ForTargets(BenefitUntilEOT(0, 2, 2)),
			Cost:    Plus(ManaCost("{1}"), TapCost(), SacrificeThis()),
			Targets: TargetCreature("target creature"),
			Effect:  pumpAndGrantFirstTarget(2, 2, "Elven Lyre — +2/+2 until end of turn"),
		}},
	})
}
