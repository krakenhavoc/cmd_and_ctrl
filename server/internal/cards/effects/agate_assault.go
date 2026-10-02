package effects

// Agate Assault — Sorcery {2}{R}:
//
//	"Choose one —
//	 • Agate Assault deals 4 damage to target creature. If that creature
//	   would die this turn, exile it instead.
//	 • Exile target artifact."
//
// The first bullet is Lava Coil's sentence: the replacement is the
// spell's, so the creature is marked whether or not the damage is dealt
// (ADR 0108 §1). The second is Murdock's Crusade's exile bullet.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "7b89b7d2-c724-4d5d-9f0b-7d3302ad1168",
		Name:         "Agate Assault",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Agate Assault deals 4 damage to target creature. If that creature would die this turn, exile it instead.",
				TargetCreature("target creature"), damageModesTargetExileIfItDies(4)),
			ModeDoing("Exile target artifact.",
				TargetPermanent("target artifact", Artifact()), exileTheModesTarget),
		),
	})
}
