package effects

// Scalding Cauldron — Artifact {1} (Throne of Eldraine, #5516):
//
//	"{3}, {T}, Sacrifice this artifact: It deals 3 damage to target creature."
//
// Bear Trap's ability without the flash. "It" is the sacrificed artifact,
// read from its last-known information.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dceb96a4-d3c5-43a2-93f6-50316e1e7a1b",
		Name:         "Scalding Cauldron",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{3}, {T}, Sacrifice this artifact: It deals 3 damage to target creature.",
			Cost:    Plus(ManaCost("{3}"), TapCost(), SacrificeThis()),
			Targets: TargetCreature("target creature"),
			Purpose: ForTargets(DamageToTarget(0, 3)),
			Effect:  sourceDealsDamageToEachLegalTarget(3),
		}},
	})
}
