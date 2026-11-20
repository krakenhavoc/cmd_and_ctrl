package effects

// Reckless Air Strike — Sorcery {R} (Core Set 2020, #5518):
//
//	"Choose one —
//	 • Reckless Air Strike deals 3 damage to target creature with flying.
//	 • Destroy target artifact."
//
// Thunderbolt's flying bullet and Qutrub Forayer's destroy bullet, each on
// its own target.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "89247308-e16f-4f60-a2a2-0b5e297c1fc1",
		Name:         "Reckless Air Strike",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Reckless Air Strike deals 3 damage to target creature with flying.",
				TargetCreature("target creature with flying", HasKeyword("flying")),
				socB20DamageToTheModeTarget(3)),
			ModeDoing("Destroy target artifact.",
				TargetPermanent("target artifact", Artifact()),
				DestroyTheModesTarget),
		),
	})
}
