package effects

// Goblin Artillery — Creature — Goblin Warrior {1}{R}{R}, 1/3 (Magic 2010, #5585):
//
//	"{T}: This creature deals 2 damage to any target and 3 damage to you."
//
// A tap ability with Spicy Oatmeal Pizza's body: 2 damage to the target and
// 3 to its controller, dealt at once. If the target is gone, the ability
// does nothing, the damage to you included (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "06db7444-0225-48c8-ab04-d69ff06e97b6",
		Name:         "Goblin Artillery",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: This creature deals 2 damage to any target and 3 damage to you.",
			Cost:    TapCost(),
			Targets: TargetAny(),
			Purpose: ForTargets(DamageToTarget(0, 2)),
			Effect:  tmtB6DamageAnyTargetAndYou(2, 3),
		}},
	})
}
