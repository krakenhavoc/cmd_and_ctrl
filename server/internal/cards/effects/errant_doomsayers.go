package effects

// Errant Doomsayers — Creature — Human Rebel {1}{W}, 1/1 (Time Spiral,
// #5596):
//
//	"{T}: Tap target creature with toughness 2 or less."
//
// Cephalid Retainer's tapper with a toughness bound, judged at announce
// and again as the ability resolves (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "82d4d70b-afb9-4ae7-8f30-16bc64dccf3e",
		Name:         "Errant Doomsayers",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: Tap target creature with toughness 2 or less.",
			Cost:    TapCost(),
			Targets: TargetCreature("target creature with toughness 2 or less", ToughnessLE(2)),
			Effect:  tapChosenPermanent,
		}},
	})
}
