package effects

// Stalking Assassin — Creature — Human Assassin {1}{U}{B}, 1/1
// (Invasion, #5616):
//
//	"{3}{U}, {T}: Tap target creature.
//	 {3}{B}, {T}: Destroy target tapped creature."
//
// Collector's Case's tap (tapChosenPermanent) and Royal Assassin's kill
// (a tappedPermanent clause and destroyFirstLegalTarget), each with a
// mana cost beside the tap. Both abilities tap the Assassin, so one turn
// can't do both; the tap means summoning sickness applies (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "12117089-67f7-4d0e-88db-82b27a4c7cdc",
		Name:         "Stalking Assassin",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{3}{U}, {T}: Tap target creature.",
				Cost:    Plus(ManaCost("{3}{U}"), TapCost()),
				Targets: TargetCreature("target creature"),
				Effect:  tapChosenPermanent,
			},
			{
				Label:   "{3}{B}, {T}: Destroy target tapped creature.",
				Cost:    Plus(ManaCost("{3}{B}"), TapCost()),
				Targets: TargetCreature("target tapped creature", tappedPermanent()),
				Effect:  destroyFirstLegalTarget,
			},
		},
	})
}
