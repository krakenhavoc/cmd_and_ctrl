package effects

// Overwhelming Surge — Instant {2}{R} (Tarkir: Dragonstorm, #4463):
//
//	"Choose one or both —
//	 • Overwhelming Surge deals 3 damage to target creature.
//	 • Destroy target noncreature artifact."
//
// Deadly Complication's "Choose one or both" (ChooseN 1..2). Each bullet
// has its own target clause, so a bullet with no legal target can't be
// chosen (CR 700.2a), and with both chosen the bullets run in printed
// order: the damage, then the destruction. A bullet whose target became
// illegal does nothing while the other still happens (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bc8a37f1-c36b-4d0f-82da-4558ff712565",
		Name:         "Overwhelming Surge",
		Completeness: CompletenessFull,
		Modes: ChooseN("Choose one or both", 1, 2,
			ModeWithPurpose(ModeDoing("Overwhelming Surge deals 3 damage to target creature.",
				TargetCreature("target creature"), DealFixedDamageToModesTarget(3)),
				ForTargets(DamageToTarget(0, 3))),
			ModeDoing("Destroy target noncreature artifact.",
				TargetPermanent("target noncreature artifact", Artifact(), Noncreature()), DestroyTheModesTarget),
		),
	})
}
