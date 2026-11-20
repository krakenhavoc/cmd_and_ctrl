package effects

// Loxodon Mystic — Creature — Elephant Cleric {3}{W}{W} (Darksteel, #5605):
//
//	"{W}, {T}: Tap target creature."
//
// Nebelgast Beguiler's tapper: the shared tapChosenPermanent body, the
// target re-checked as the ability resolves (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f3e22884-a9ca-4479-b400-df1c1d2fd153",
		Name:         "Loxodon Mystic",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{W}, {T}: Tap target creature.",
			Cost:    Plus(ManaCost("{W}"), TapCost()),
			Targets: TargetCreature("target creature"),
			Effect:  tapChosenPermanent,
		}},
	})
}
