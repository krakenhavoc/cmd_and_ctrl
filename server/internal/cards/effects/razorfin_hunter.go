package effects

// Razorfin Hunter — Creature — Merfolk Goblin {U}{R}, 1/1 (Apocalypse,
// #5614):
//
//	"{T}: This creature deals 1 damage to any target."
//
// Frostwielder's tap ping on a two-mana body. The {T} cost waits out
// summoning sickness (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "45baa5d1-79aa-4583-b938-0d5feaa5aa0c",
		Name:         "Razorfin Hunter",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: This creature deals 1 damage to any target.",
			Cost:    TapCost(),
			Targets: TargetAny(),
			Purpose: ForTargets(DamageToTarget(0, 1)),
			Effect:  sourceDealsDamageToEachLegalTarget(1),
		}},
	})
}
