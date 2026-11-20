package effects

// Anointer of Champions — Creature — Human Cleric {W}, 1/1 (Magic Origins, #5549):
//
//	"{T}: Target attacking creature gets +1/+1 until end of turn."
//
// Infantry Veteran's tap: the target must be an attacking creature at
// activation and at resolution.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f74ef5b0-9222-43d0-a919-aaf28dc21b59",
		Name:         "Anointer of Champions",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: Target attacking creature gets +1/+1 until end of turn.",
			Purpose: ForTargets(BenefitUntilEOT(0, 1, 1)),
			Cost:    TapCost(),
			Targets: TargetCreature("target attacking creature", AttackingCreature()),
			Effect:  pumpAndGrantFirstTarget(1, 1, "Anointer of Champions — +1/+1 until end of turn"),
		}},
	})
}
