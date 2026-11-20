package effects

// Chandra's Magmutt — Creature — Elemental Dog {1}{R}, 2/2 (Core Set 2021, #5511):
//
//	"{T}: This creature deals 1 damage to target player or planeswalker."
//
// Goblin Fireslinger's tap ping (sourceDealsOneToFirstTarget) with its
// "target player or planeswalker" clause.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a2c4537e-70a6-416d-9e8b-f549ad535147",
		Name:         "Chandra's Magmutt",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: This creature deals 1 damage to target player or planeswalker.",
			Cost:    TapCost(),
			Targets: targetPlayerOrPlaneswalker(),
			Purpose: ForTargets(DamageToTarget(0, 1)),
			Effect:  sourceDealsOneToFirstTarget,
		}},
	})
}
