package effects

// Pinecone Strike — Instant {1}{R}:
//
//	"Choose one or both —
//	 • Pinecone Strike deals 3 damage to target creature. If that
//	   creature would die this turn, exile it instead.
//	 • Destroy target artifact token."
//
// Each bullet reads its own target group, and the chosen bullets
// resolve in printed order (CR 608.2c). The creature is marked whether
// or not the damage is dealt (ADR 0108 §1).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "961e3023-39ea-4141-99b6-738280a2815d",
		Name:         "Pinecone Strike",
		Completeness: CompletenessFull,
		Modes: ChooseN("Choose one or both", 1, 2,
			ModeWithPurpose(ModeDoing("Pinecone Strike deals 3 damage to target creature. If that creature would die this turn, exile it instead.",
				TargetCreature("target creature"), damageModesTargetExileIfItDies(3)), ForTargets(DamageToTarget(0, 3))),
			ModeDoing("Destroy target artifact token.",
				TargetPermanent("target artifact token", Artifact(), IsTokenPredicate()), DestroyTheModesTarget),
		),
	})
}
