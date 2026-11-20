package effects

// Deface — Sorcery {R} (Ravnica Allegiance, #5521):
//
//	"Choose one —
//	 • Destroy target artifact.
//	 • Destroy target creature with defender."
//
// Choose one (CR 700.2): Bant Charm's artifact bullet, and a creature
// bullet whose target must have defender (a layer-6 keyword, so a granted
// defender counts).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "390290b3-77cc-4180-8e6c-9cb808f6ac7a",
		Name:         "Deface",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Destroy target artifact.",
				TargetPermanent("target artifact", Artifact()),
				DestroyTheModesTarget),
			ModeDoing("Destroy target creature with defender.",
				TargetCreature("target creature with defender", HasKeyword("defender")),
				DestroyTheModesTarget),
		),
	})
}
