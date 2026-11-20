package effects

// Checkpoint Officer — Creature — Human Soldier {1}{W}, 1/2 (Ikoria: Lair of Behemoths, #5512):
//
//	"{1}{W}, {T}: Tap target creature."
//
// Celestial Enforcer's tapper without its condition (tapChosenPermanent).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "03e25120-b4e4-409e-8dbc-7574ea4c78d6",
		Name:         "Checkpoint Officer",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}{W}, {T}: Tap target creature.",
			Cost:    Plus(ManaCost("{1}{W}"), TapCost()),
			Targets: TargetCreature("target creature"),
			Effect:  tapChosenPermanent,
		}},
	})
}
