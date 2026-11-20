package effects

// Gideon's Lawkeeper — Creature — Human Soldier {W}, 1/1 (Magic 2012,
// #5573):
//
//	"{W}, {T}: Tap target creature."
//
// Nebelgast Beguiler's text: {W}, {T}: tap target creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bae86bb0-5a97-4f6a-8a3f-4e6848bfacc0",
		Name:         "Gideon's Lawkeeper",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{W}, {T}: Tap target creature.",
			Cost:    Plus(ManaCost("{W}"), TapCost()),
			Targets: TargetCreature("target creature"),
			Effect:  tapChosenPermanent,
		}},
	})
}
