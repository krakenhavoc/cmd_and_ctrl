package effects

// Suplex — Sorcery {1}{R}:
//
//	"Choose one —
//	 • Suplex deals 3 damage to target creature. If that creature would
//	   die this turn, exile it instead.
//	 • Exile target artifact."
//
// Agate Assault's two bullets for 3 damage: the creature is marked
// whether or not the damage is dealt (ADR 0108 §1).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "2d1bbeda-2e81-4aaa-9494-094ec3dd6c3b",
		Name:         "Suplex",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeWithPurpose(ModeDoing("Suplex deals 3 damage to target creature. If that creature would die this turn, exile it instead.",
				TargetCreature("target creature"), damageModesTargetExileIfItDies(3)), ForTargets(DamageToTarget(0, 3))),
			ModeDoing("Exile target artifact.",
				TargetPermanent("target artifact", Artifact()), exileTheModesTarget),
		),
	})
}
