package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blood Funnel — Enchantment {1}{B} (Ravnica: City of Guilds):
//
//	"Noncreature spells you cast cost {2} less to cast.
//	 Whenever you cast a noncreature spell, counter that spell unless you sacrifice
//	 a creature."
//
// The reduction is generic-only and stops at zero (CR 601.2f), over the
// controller's own noncreature spells. The trigger goes on the stack above the
// spell; as it resolves, its caster is asked to sacrifice a creature of their
// choice, and a "no", or a "yes" with no creature to sacrifice (CR 118.3),
// counters the spell. The spell is judged a noncreature spell as it sits on the
// stack, so a Blood Funnel that is cast itself is not taxed by its own text.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "81dd22f7-f6c6-4cf2-9744-b1d918dbcf9a",
		Name:         "Blood Funnel",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.CostsLess(2, "Noncreature spells you cast cost {2} less to cast.", effects.YourSpell(), effects.NoncreatureSpell()),
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.YouCast(effects.Noncreature()),
				"Blood Funnel — counter that spell unless you sacrifice a creature",
				effects.OldB579CounterUnlessYouSacrificeACreature("Blood Funnel")),
		},
	})
}
