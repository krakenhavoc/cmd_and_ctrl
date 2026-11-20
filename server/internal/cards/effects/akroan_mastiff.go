package effects

// Akroan Mastiff — Creature — Dog {3}{W}, 2/2 (Journey into Nyx, #5558):
//
//	"{W}, {T}: Tap target creature."
//
// Nebelgast Beguiler's tapper.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6039b9b7-e49a-4ea3-887c-a17dd502d63f",
		Name:         "Akroan Mastiff",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{W}, {T}: Tap target creature.",
			Cost:    Plus(ManaCost("{W}"), TapCost()),
			Targets: TargetCreature("target creature"),
			Effect:  tapChosenPermanent,
		}},
	})
}
