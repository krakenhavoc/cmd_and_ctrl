package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rain of Thorns — Sorcery {4}{G}{G} (Commander 2017, #5533):
//
//	"Choose one or more —
//	 • Destroy target artifact.
//	 • Destroy target enchantment.
//	 • Destroy target land."
//
// Kill! Maim! Burn!'s modal destroy: each chosen bullet announces its own
// target group (CR 700.2d), all destroyed as the spell resolves, and a
// bullet whose target has become illegal does nothing while the others
// still happen (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "363b0504-27bf-40df-96ec-9e3cb1e47588",
		Name:         "Rain of Thorns",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOneOrMore(
			effects.ModeDoing("Destroy target artifact.", effects.TargetPermanent("target artifact", effects.Artifact()), effects.DestroyTheModesTarget),
			effects.ModeDoing("Destroy target enchantment.", effects.TargetPermanent("target enchantment", effects.Enchantment()), effects.DestroyTheModesTarget),
			effects.ModeDoing("Destroy target land.", effects.TargetPermanent("target land", effects.Land()), effects.DestroyTheModesTarget),
		),
	})
}
