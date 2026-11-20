package effects

// Thundering Rebuke — Sorcery {1}{R} (Zendikar Rising, #5508):
//
//	"Thundering Rebuke deals 4 damage to target creature or planeswalker."
//
// Sear's text at sorcery speed (the speed rides the type line).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bb4d9770-f230-4413-a7a4-bb1f1303661a",
		Name:         "Thundering Rebuke",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
		Purpose:      ForTargets(DamageToTarget(0, 4)),
		OnResolve:    damageToFirstTarget(4),
	})
}
