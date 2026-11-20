package effects

// Orcish Vandal — Creature — Orc Warrior {1}{R}, 1/1 (Dominaria, #5530):
//
//	"{T}, Sacrifice an artifact: This creature deals 2 damage to any target."
//
// Orcish Mechanics' ability: tap it and sacrifice an artifact for 2
// damage to any target.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "933a1811-7e1a-4173-b1c9-34b2c858ba68",
		Name:         "Orcish Vandal",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}, Sacrifice an artifact: This creature deals 2 damage to any target.",
			Cost:    Plus(TapCost(), SacrificeN(1, "an artifact", Artifact())),
			Targets: TargetAny(),
			Purpose: ForTargets(DamageToTarget(0, 2)),
			Effect:  DealDamageToTheTarget(2),
		}},
	})
}
