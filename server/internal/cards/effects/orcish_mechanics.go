package effects

// Orcish Mechanics — Creature — Orc {2}{R}, 1/1 (Antiquities, #5640):
//
//	"{T}, Sacrifice an artifact: This creature deals 2 damage to any
//	 target."
//
// Pia and Kiran Nalaar's sacrifice ping with a {T} in place of the mana:
// the artifact is paid at announce, the target is any creature, player,
// planeswalker or battle, and a target that left in response leaves the
// ability to do nothing (DealDamageToTheTarget).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0ec58835-de2d-4064-89a9-f92db80bc276",
		Name:         "Orcish Mechanics",
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
