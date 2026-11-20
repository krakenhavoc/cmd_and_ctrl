package effects

// Viridian Scout — Creature — Elf Warrior Scout {3}{G}, 1/2 (Fifth Dawn, #5604):
//
//	"{2}{G}, Sacrifice this creature: It deals 2 damage to target
//	 creature with flying."
//
// Barbarian Lunatic's sacrifice-and-shoot, limited to a flier (Centaur
// Archer's clause). The Scout is sacrificed as a cost, so the damage comes
// from it as it last existed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c5b4bc25-0f32-4e1d-b6fe-f122140f2207",
		Name:         "Viridian Scout",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}{G}, Sacrifice this creature: It deals 2 damage to target creature with flying.",
			Purpose: ForTargets(DamageToTarget(0, 2)),
			Cost:    Plus(ManaCost("{2}{G}"), SacrificeThis()),
			Targets: TargetCreature("target creature with flying", HasKeyword("flying")),
			Effect:  DealDamageToTheTarget(2),
		}},
	})
}
