package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mockery of Nature — Creature — Eldrazi Beast {9}, 6/5:
//
//	"Emerge {7}{G} (You may cast this spell by sacrificing a creature and
//	 paying the emerge cost reduced by that creature's mana value.)
//	 When you cast this spell, you may destroy target artifact or
//	 enchantment."
//
// Emerge is the shared alternative cost (ADR 0135 §4). The cast trigger
// is a "you may": the caster is asked as it triggers and on a yes picks
// the artifact or enchantment, which is destroyed above the spell even if
// the Mockery is countered. A target gone by then is left alone (CR
// 608.2b).
//
// No simplification.
func init() {
	cast := effects.WhenYouCastThisSpell("Mockery of Nature — destroy target artifact or enchantment", effects.DestroyChosenPermanent)
	cast.Targets = effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment()))
	effects.Register(effects.Spec{
		OracleID:         "5da5240e-5d87-494f-aec9-64b7a3f0d935",
		Name:             "Mockery of Nature",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Emerge("{7}{G}")},
		Triggered: []game.TriggeredAbility{
			effects.Optional(cast, "Mockery of Nature — destroy an artifact or enchantment?"),
		},
	})
}
