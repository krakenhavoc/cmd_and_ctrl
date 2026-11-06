package effects

// Frostbridge Guard — Creature — Elemental Soldier {1}{W}, 2/2 (Wilds of
// Eldraine, #4471):
//
//	"{2}{W}, {T}: Tap target creature."
//
// Kytheon's Irregulars' tap ability at {2}{W} plus {T}; the {T} in the
// cost means a summoning-sick Guard can't activate it (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "515c1604-59f0-45b4-91be-d2f20fdd3e1e",
		Name:         "Frostbridge Guard",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}{W}, {T}: Tap target creature.",
			Cost:    Plus(ManaCost("{2}{W}"), TapCost()),
			Targets: TargetCreature("target creature"),
			Effect:  tapChosenPermanent,
		}},
	})
}
