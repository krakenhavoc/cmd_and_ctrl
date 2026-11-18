package effects

// Gavony Trapper — Creature — Human Soldier {W}, 0/2 (Innistrad: Midnight Hunt, S82):
//
//	"{2}, {T}: Tap target creature."
//
// Arashin Sunshield's tapper with a generic cost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d4da20c7-ae4a-44c6-8f03-9e719ccee8c9",
		Name:         "Gavony Trapper",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}, {T}: Tap target creature.",
			Cost:    Plus(ManaCost("{2}"), TapCost()),
			Targets: TargetCreature("target creature"),
			Effect:  tapChosenPermanent,
		}},
	})
}
