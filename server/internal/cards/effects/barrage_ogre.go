package effects

// Barrage Ogre — Creature — Ogre Warrior {3}{R}{R}, 3/3 (Scars of Mirrodin, #5577):
//
//	"{T}, Sacrifice an artifact: This creature deals 2 damage to any target."
//
// Orcish Mechanics' ability: the artifact is sacrificed as a cost
// (CR 602.2b), so it is gone before the target is dealt damage.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d556c71e-ce16-4229-880b-744790f93797",
		Name:         "Barrage Ogre",
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
