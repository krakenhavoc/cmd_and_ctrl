package effects

// Zulaport Chainmage — Creature — Human Shaman Ally {3}{B}, 4/2 (Oath of the Gatewatch, #5545):
//
//	"Cohort — {T}, Tap an untapped Ally you control: Target opponent
//	 loses 2 life."
//
// Cohort is an ability word (CR 207.2c): the cost is this creature's {T}
// plus tapping an untapped Ally you control (the tap-others cost,
// game/tap_others_cost.go), which is not the {T} symbol, so an Ally that
// arrived this turn may pay it. The creature can't pay both halves (its
// own {T} has tapped it), so the clause excludes it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "017f4471-db5a-4ad0-97b3-b1f12b52578a",
		Name:         "Zulaport Chainmage",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Cohort — {T}, Tap an untapped Ally you control: Target opponent loses 2 life.",
			Cost:    ogwCohortCost(),
			Targets: TargetPlayer("target opponent", Opponent()),
			Effect:  Do(TargetPlayerLosesLife{N: 2}),
		}},
	})
}
