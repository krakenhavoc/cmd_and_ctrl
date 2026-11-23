package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "5c6c8fe7-3520-423b-a224-1c0af516871a",
		Name:         "Casualties of War",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOneOrMore(
			effects.ModeDoing("Destroy target artifact.", effects.TargetPermanent("target artifact", effects.Artifact()), effects.DestroyTheModesTarget),
			effects.ModeDoing("Destroy target creature.", effects.TargetCreature("target creature"), effects.DestroyTheModesTarget),
			effects.ModeDoing("Destroy target enchantment.", effects.TargetPermanent("target enchantment", effects.Enchantment()), effects.DestroyTheModesTarget),
			effects.ModeDoing("Destroy target land.", effects.TargetPermanent("target land", effects.Land()), effects.DestroyTheModesTarget),
			effects.ModeDoing("Destroy target planeswalker.", effects.TargetPermanent("target planeswalker", effects.Planeswalker()), effects.DestroyTheModesTarget),
		),
	})
}
