package effects

// Scrap Compactor — Artifact {1} (Aetherdrift, #4464):
//
//	"{3}, {T}, Sacrifice this artifact: It deals 3 damage to target
//	 creature.
//	 {6}, {T}, Sacrifice this artifact: Destroy target creature or
//	 Vehicle."
//
// Bear Trap's sacrifice-for-damage row (the sacrificed Compactor deals
// the damage on its last-known information, CR 113.7a) beside Cathar
// Commando's sacrifice-and-destroy row with Spin Out's "creature or
// Vehicle" target, so an uncrewed Vehicle is a legal pick.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7f113875-a411-463b-8b4d-91d1dee95c2a",
		Name:         "Scrap Compactor",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{3}, {T}, Sacrifice this artifact: It deals 3 damage to target creature.",
				Cost:    Plus(ManaCost("{3}"), TapCost(), SacrificeThis()),
				Targets: TargetCreature("target creature"),
				Purpose: ForTargets(DamageToTarget(0, 3)),
				Effect:  sourceDealsDamageToEachLegalTarget(3),
			},
			{
				Label:   "{6}, {T}, Sacrifice this artifact: Destroy target creature or Vehicle.",
				Cost:    Plus(ManaCost("{6}"), TapCost(), SacrificeThis()),
				Targets: TargetPermanent("target creature or Vehicle", Or(Creature(), HasSubtype("Vehicle"))),
				Effect:  destroyChosenPermanent,
			},
		},
	})
}
