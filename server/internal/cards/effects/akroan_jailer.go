package effects

// Akroan Jailer — Creature — Human Soldier {W}, 1/1 (Magic Origins, #5549):
//
//	"{2}{W}, {T}: Tap target creature."
//
// Frostbridge Guard's activated tap.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e3ad2c77-2a41-4a4a-a6bf-e542d0162b88",
		Name:         "Akroan Jailer",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}{W}, {T}: Tap target creature.",
			Cost:    Plus(ManaCost("{2}{W}"), TapCost()),
			Targets: TargetCreature("target creature"),
			Effect:  tapChosenPermanent,
		}},
	})
}
