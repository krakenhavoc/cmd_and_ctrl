package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sage's Reverie — Enchantment — Aura {3}{W}:
//
//	"Enchant creature
//	 When this Aura enters, draw a card for each Aura you control
//	 that's attached to a creature.
//	 Enchanted creature gets +1/+1 for each Aura you control that's
//	 attached to a creature."
//
// Both counts are the same question asked at different times: the
// draw is counted as the trigger resolves, the bonus on every layer
// recompute. The Aura itself counts (it is attached to a creature by
// the time its own trigger is put on the stack).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bd92cacd-c7d8-43a2-a261-ea4754fefdc5",
		Name:         "Sage's Reverie",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Static: []game.StaticAbility{
			effects.PumpAttachedPer(1, 1, effects.AurasYouControlOnCreatures),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Sage's Reverie — draw a card for each Aura you control attached to a creature",
				func(g *game.Game, item *game.StackItem) error {
					src, ok := g.LookupCardForEffect(item.SourceCardID)
					if !ok {
						return nil
					}
					n := effects.AurasYouControlOnCreatures(g, &src)
					if n <= 0 {
						return nil
					}
					return effects.DrawCards{Player: item.Controller, N: n}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
