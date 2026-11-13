package effects

// Icatian Scout — Creature — Human Soldier Scout {W}, 1/1 (Fallen
// Empires, #5637):
//
//	"{1}, {T}: Target creature gains first strike until end of turn."
//
// A layer-6 grant on the still-legal target (CR 608.2b), ending at
// cleanup. The {T} in the cost keeps it off the turn the Scout arrives.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b1a9f49a-b0e2-481f-806d-0bc321517a4f",
		Name:         "Icatian Scout",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}, {T}: Target creature gains first strike until end of turn.",
			Purpose: ForTargets(BenefitUntilEOT(0, 0, 0, "first strike")),
			Cost:    Plus(ManaCost("{1}"), TapCost()),
			Targets: TargetCreature("target creature"),
			Effect:  pumpAndGrantFirstTarget(0, 0, "Icatian Scout — first strike until end of turn", "first strike"),
		}},
	})
}
