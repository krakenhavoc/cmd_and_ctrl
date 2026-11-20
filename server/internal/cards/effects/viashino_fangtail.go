package effects

// Viashino Fangtail — Creature — Lizard Warrior {2}{R}{R}, 3/3 (Ravnica:
// City of Guilds, #5600):
//
//	"{T}: This creature deals 1 damage to any target."
//
// Thornwind Faeries' tap ping: the creature is the source of the damage.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5140cd2f-af30-4405-b5e4-996ef6349f59",
		Name:         "Viashino Fangtail",
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
