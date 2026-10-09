package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	cast := WhenYouCastThisSpell("Mockery of Nature — destroy target artifact or enchantment", destroyChosenPermanent)
	cast.Targets = TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment()))
	Register(Spec{
		OracleID:         "5da5240e-5d87-494f-aec9-64b7a3f0d935",
		Name:             "Mockery of Nature",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Emerge("{7}{G}")},
		Triggered: []game.TriggeredAbility{
			Optional(cast, "Mockery of Nature — destroy an artifact or enchantment?"),
		},
	})
}
