package effects

// Goblin Cratermaker — Creature — Goblin Warrior {1}{R} (The Hobbit Commander), 2/2:
//
//	"{1}, Sacrifice this creature: Choose one —
//	 • This creature deals 2 damage to target creature.
//	 • Destroy target colorless nonland permanent."
//
// #4454, S76. A modal activated ability is one ability per bullet with
// the same cost (Insidious Fungus's shape): the mode is chosen at
// activation (CR 700.2) and the client's ability menu is the picker.
// The damage is dealt by the Cratermaker as it last existed (CR 608.2h).
//
// No simplification.
func init() {
	cost := Plus(ManaCost("{1}"), SacrificeThis())
	Register(Spec{
		OracleID:     "1962d9b3-f057-4d2c-9c5b-386d9a66148a",
		Name:         "Goblin Cratermaker",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{1}, Sacrifice this creature: This creature deals 2 damage to target creature.",
				Cost:    cost,
				Targets: TargetCreature("target creature"),
				Purpose: ForTargets(DamageToTarget(0, 2)),
				Effect:  sourceDealsDamageToEachLegalTarget(2),
			},
			{
				Label:   "{1}, Sacrifice this creature: Destroy target colorless nonland permanent.",
				Cost:    cost,
				Targets: TargetPermanent("target colorless nonland permanent", And(Colorless(), Nonland())),
				Effect:  destroyFirstLegalTarget,
			},
		},
	})
}
