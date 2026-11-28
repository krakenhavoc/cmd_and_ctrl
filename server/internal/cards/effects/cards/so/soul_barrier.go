package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Soul Barrier — Enchantment {2}{U} (Ice Age, S82 OLD batch 840, #5636):
//
//	"Whenever an opponent casts a creature spell, this enchantment deals 2 damage to
//	 that player unless they pay {2}."
//
// Mystic Remora's tax: the trigger is on the cast of an opponent's creature spell
// (read off the spell on the stack), and the caster, not the enchantment's
// controller, is asked to pay {2} as it resolves; a decline has this enchantment
// deal 2 damage to them. The damage is the enchantment's, by its last-known
// information if it has left.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3aab67bb-f287-4aa6-a3c5-e850b22dfff6",
		Name:         "Soul Barrier",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.AnOpponentCast(effects.Creature()),
				"Soul Barrier — 2 damage to that player unless they pay {2}",
				func(g *game.Game, item *game.StackItem) error {
					return effects.PayUnless{
						Chooser:  item.Trigger.Event.Actor,
						Cost:     "{2}",
						Question: "Soul Barrier — pay {2} or take 2 damage?",
						Else:     soulBarrierDamage,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// soulBarrierDamage is the decline: 2 damage to the caster. Keyed (#4264); an
// on-disk identity.
var soulBarrierDamage = effects.PayUnlessElse("pay-unless/soul-barrier-damage", func(ctx *effects.Context, r game.PayUnlessAnswer) error {
	return effects.DealDamage{Source: r.Source, Target: r.Chooser, Amount: 2}.Apply(ctx)
})
