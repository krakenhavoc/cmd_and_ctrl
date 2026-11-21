package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Impending Doom — Enchantment — Aura {2}{R} (Jumpstart 2022):
//
//	"Enchant creature
//	 Enchanted creature gets +3/+3 and attacks each combat if able.
//	 When enchanted creature dies, this Aura deals 3 damage to that creature's
//	 controller."
//
// Furor of the Bitten's two statics. The dies trigger is harvested while the
// Aura is still attached (WhenEnchantedCreatureDies) and hits the creature's
// controller as it last was (CR 603.10a), so a creature stolen before it died
// costs its thief the life; the Aura is the damage source.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "94366283-3b28-4d6a-96e7-c04e5aead5a6",
		Name:         "Impending Doom",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttached(3, 3),
			effects.AttacksEachCombatWhere(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return source.AttachedTo.Kind == game.TargetCard && source.AttachedTo.ID == target.InstanceID
			}),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenEnchantedCreatureDies("Impending Doom — 3 damage to that creature's controller",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					ev := item.Trigger.Event
					dead, ok := g.LookupCardForEffect(ev.CardID)
					if !ok {
						return nil
					}
					return effects.DealDamage{Source: item.SourceCardID, Target: effects.LeftUnderControlOf(ev, dead), Amount: 3}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
