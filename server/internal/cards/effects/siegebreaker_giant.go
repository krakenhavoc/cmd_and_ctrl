package effects

// Siegebreaker Giant — Creature — Giant Warrior {3}{R}{R}, 6/3
// (Core Set 2019, #5527):
//
//	"Trample (This creature can deal excess combat damage to the player or planeswalker it's attacking.)
//	{3}{R}: Target creature can't block this turn."
//
// Trample, and Chainwhip Cyclops's ability: one target creature can't
// block this turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "fcd281e3-f336-449f-a887-4b98dc14dfe7",
		Name:            "Siegebreaker Giant",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Activated: []ActivatedAbility{{
			Label:   "{3}{R}: Target creature can't block this turn.",
			Cost:    ManaCost("{3}{R}"),
			Targets: TargetCreature("target creature"),
			Purpose: ForTargets(CantBlockUntilEOT(0, 0, 0)),
			Effect:  targetCreatureCantBlockThisTurn("Siegebreaker Giant — target creature can't block this turn"),
		}},
	})
}
