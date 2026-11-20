package effects

// Zada's Commando — Creature — Goblin Archer Ally {1}{R}, 2/1 (Oath of the Gatewatch, #5545):
//
//	"First strike
//	 Cohort — {T}, Tap an untapped Ally you control: This creature
//	 deals 1 damage to target opponent or planeswalker."
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
		OracleID:        "d57d1988-5962-4df9-8342-cf04233fd6c7",
		Name:            "Zada's Commando",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Activated: []ActivatedAbility{{
			Label:   "Cohort — {T}, Tap an untapped Ally you control: This creature deals 1 damage to target opponent or planeswalker.",
			Purpose: ForTargets(DamageToTarget(0, 1)),
			Cost:    ogwCohortCost(),
			Targets: targetOpponentOrPlaneswalker(),
			Effect:  DealDamageToTheTarget(1),
		}},
	})
}
