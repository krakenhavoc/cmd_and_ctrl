package effects

// Barrage of Expendables — Enchantment {R} (Magic 2014, #5563):
//
//	"{R}, Sacrifice a creature: This enchantment deals 1 damage to any target."
//
// Goblin Bombardment at {R} a shot: the enchantment is the damage's source.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "da7600f6-5d4c-46bf-936a-9e0f0d508e19",
		Name:         "Barrage of Expendables",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{R}, Sacrifice a creature: This enchantment deals 1 damage to any target.",
			Cost:    Plus(ManaCost("{R}"), SacrificeACreature()),
			Targets: TargetAny(),
			Purpose: ForTargets(DamageToTarget(0, 1)),
			Effect:  DamageToFirstTarget(1),
		}},
	})
}
