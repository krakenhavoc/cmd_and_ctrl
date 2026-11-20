package effects

// Goblin Tunneler — Creature — Goblin Rogue {1}{R}, 1/1 (Rise of the
// Eldrazi, #5581):
//
//	"{T}: Target creature with power 2 or less can't be blocked this turn."
//
// Crafty Pathmage's ability. The power is checked as the target is chosen
// and again as the ability resolves; a creature pumped past 2 afterwards
// stays unblockable.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1005bc16-8a6a-476a-a647-4ed40d2e96de",
		Name:         "Goblin Tunneler",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: Target creature with power 2 or less can't be blocked this turn.",
			Purpose: ForTargets(CantBeBlockedUntilEOT(0, 0, 0)),
			Cost:    TapCost(),
			Targets: TargetCreature("target creature with power 2 or less", PowerLE(2)),
			Effect:  targetCantBeBlockedThisTurn("Goblin Tunneler"),
		}},
	})
}
