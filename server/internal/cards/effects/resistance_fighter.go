package effects

// Resistance Fighter — Creature — Human Soldier {W}:
//
//	"Sacrifice this creature: Prevent all combat damage target creature would deal this turn."
//
// ADR 0108 §7 (#1904, Delivery PR 7b): the shield's source is the
// targeted creature, pinned as the ability resolves (CR 400.7).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "a95a2d1c-70a9-4eb2-ae30-a07fab26a9c6",
		Name:         "Resistance Fighter",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{shieldAgainstTargetsRow(
			"Sacrifice this creature: Prevent all combat damage target creature would deal this turn.",
			SacrificeThis(), TargetCreature("target creature"), true)},
	})
}
