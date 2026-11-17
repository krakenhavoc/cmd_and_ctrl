package effects

// Expendable Troops — Creature — Human Soldier {1}{W}, 2/1 (Urza's
// Legacy, #5623):
//
//	"{T}, Sacrifice this creature: It deals 2 damage to target
//	 attacking or blocking creature."
//
// Divebomber Griffin's tap-and-sacrifice row at 2 damage, aimed at a creature
// that is attacking or blocking as it is announced and still is at
// resolution (CR 608.2b). The Troops are the source of the damage though
// they are already in the graveyard.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "70ebd18a-c878-4a93-b7c4-3f7a2aa65f5e",
		Name:         "Expendable Troops",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}, Sacrifice this creature: It deals 2 damage to target attacking or blocking creature.",
			Cost:    Plus(TapCost(), SacrificeThis()),
			Targets: TargetCreature("target attacking or blocking creature", AttackingOrBlocking()),
			Purpose: ForTargets(DamageToTarget(0, 2)),
			Effect:  b33DamageChosenTargetFromSource(2),
		}},
	})
}
