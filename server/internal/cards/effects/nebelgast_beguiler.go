package effects

// Nebelgast Beguiler — Creature — Spirit {4}{W}, 2/5 (Innistrad: Crimson
// Vow):
//
//	"{W}, {T}: Tap target creature."
//
// Arashin Sunshield's tap ability.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b1e76ce9-6f39-419c-99e3-ede2dfa4858c",
		Name:         "Nebelgast Beguiler",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{W}, {T}: Tap target creature.",
			Cost:    Plus(ManaCost("{W}"), TapCost()),
			Targets: TargetCreature("target creature"),
			Effect:  tapChosenPermanent,
		}},
	})
}
