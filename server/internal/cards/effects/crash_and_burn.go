package effects

// Crash and Burn — Instant {3}{R} (Aetherdrift, #4464):
//
//	"Choose one —
//	 • Destroy target Vehicle.
//	 • Crash and Burn deals 6 damage to target creature or planeswalker."
//
// Destroy Evil's modal shape. The first bullet's Vehicle may be crewed or
// not; the second is Fulminous Forte's fixed damage to a creature or
// planeswalker. Each bullet's target is re-checked as it resolves
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "38405a37-921a-49ab-a95b-28eb8c26474a",
		Name:         "Crash and Burn",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Destroy target Vehicle.",
				TargetPermanent("target Vehicle", HasSubtype("Vehicle")),
				DestroyTheModesTarget),
			ModeWithPurpose(ModeDoing("Crash and Burn deals 6 damage to target creature or planeswalker.",
				TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
				DealFixedDamageToModesTarget(6)), ForTargets(DamageToTarget(0, 6))),
		),
	})
}
