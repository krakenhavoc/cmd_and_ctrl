package effects

// Goblin Smuggler — Creature — Goblin Rogue {2}{R}, 2/2 (Foundations,
// #4465):
//
//	"Haste
//	 {T}: Another target creature with power 2 or less can't be blocked
//	 this turn."
//
// Haste rides PrintedKeywords, so the tap ability is usable the turn the
// Goblin arrives. The target is any other creature, whoever controls it
// (Another sets ExcludeSource); the power limit is checked as the ability
// is activated and again as it resolves (CR 608.2b), so a creature that
// grew in response is skipped. The body is Access Tunnel's
// (targetCantBeBlockedThisTurn).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "eae15f7d-1f6e-4dfc-afea-7698b93cd145",
		Name:            "Goblin Smuggler",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Activated: []ActivatedAbility{{
			Label:   "{T}: Another target creature with power 2 or less can't be blocked this turn.",
			Purpose: ForTargets(CantBeBlockedUntilEOT(0, 0, 0)),
			Cost:    TapCost(),
			Targets: Another(TargetCreature("another target creature with power 2 or less", PowerLE(2))),
			Effect:  targetCantBeBlockedThisTurn("Goblin Smuggler"),
		}},
	})
}
