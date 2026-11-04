package effects

// Casualties of War — Sorcery {2}{B}{B}{G}{G} (Secrets of Strixhaven
// Commander, #4834):
//
//	"Choose one or more —
//	 • Destroy target artifact.
//	 • Destroy target creature.
//	 • Destroy target enchantment.
//	 • Destroy target land.
//	 • Destroy target planeswalker."
//
// ChooseOneOrMore over five targeted modes, each its own target group,
// destroyed by DestroyTheModesTarget (Collective Resistance's bullets).
// Each mode's target is checked on its own as the spell resolves, so a
// target that became illegal in response is skipped and the rest are
// still destroyed (CR 608.2b). One permanent of two types (an artifact
// creature) may be the target of two modes, as printed.
func init() {
	Register(Spec{
		OracleID:     "5c6c8fe7-3520-423b-a224-1c0af516871a",
		Name:         "Casualties of War",
		Completeness: CompletenessFull,
		Modes: ChooseOneOrMore(
			ModeDoing("Destroy target artifact.", TargetPermanent("target artifact", Artifact()), DestroyTheModesTarget),
			ModeDoing("Destroy target creature.", TargetCreature("target creature"), DestroyTheModesTarget),
			ModeDoing("Destroy target enchantment.", TargetPermanent("target enchantment", Enchantment()), DestroyTheModesTarget),
			ModeDoing("Destroy target land.", TargetPermanent("target land", Land()), DestroyTheModesTarget),
			ModeDoing("Destroy target planeswalker.", TargetPermanent("target planeswalker", Planeswalker()), DestroyTheModesTarget),
		),
	})
}
