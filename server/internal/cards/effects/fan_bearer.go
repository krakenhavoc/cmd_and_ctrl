package effects

// Fan Bearer — Creature — Zombie {W}, 1/2
// (Amonkhet, #5535):
//
//	"{2}, {T}: Tap target creature."
//
// Gavony Trapper's ability.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d2e7198a-0eb5-4d30-973f-4b2d9617436f",
		Name:         "Fan Bearer",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}, {T}: Tap target creature.",
			Cost:    Plus(ManaCost("{2}"), TapCost()),
			Targets: TargetCreature("target creature"),
			Effect:  tapChosenPermanent,
		}},
	})
}
