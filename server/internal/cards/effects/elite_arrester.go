package effects

// Elite Arrester — Creature — Human Soldier {W}, 0/3 (Ravnica Allegiance,
// #5521):
//
//	"{1}{U}, {T}: Tap target creature."
//
// Arashin Sunshield's tapper at a different cost: tapChosenPermanent.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f25af8de-3edb-400b-9b6c-ec446a8314fa",
		Name:         "Elite Arrester",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}{U}, {T}: Tap target creature.",
			Cost:    Plus(ManaCost("{1}{U}"), TapCost()),
			Targets: TargetCreature("target creature"),
			Effect:  tapChosenPermanent,
		}},
	})
}
