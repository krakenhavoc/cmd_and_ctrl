package cardslu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lust for War — Enchantment — Aura {2}{R} (Rise of the Eldrazi, S82 OLD
// batch 484, #5581):
//
//	"Enchant creature
//	 Whenever enchanted creature becomes tapped, this Aura deals 3 damage to
//	 that creature's controller.
//	 Enchanted creature attacks each combat if able."
//
// Brink of Disaster's tap watch (a tap by any cause, an attack included) and
// Fealty to the Realm's attack requirement. The damage is the Aura's, dealt
// to whoever controls the creature as the trigger resolves, or as it last
// existed if it has gone (CR 608.2h).
//
// No simplification.
func init() {
	requirement := effects.AttacksEachCombatWhere(effects.AttachedToSource)
	requirement.Label = "Enchanted creature attacks each combat if able."
	effects.Register(effects.Spec{
		OracleID:     "0ba206f3-c3f3-4e53-b072-e5e70feaa00d",
		Name:         "Lust for War",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static:       []game.StaticAbility{requirement},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventTapCard, game.EventAttack},
			AppliesTo: effects.EnchantedPermanentBecameTapped,
			Key:       "Lust for War — 3 damage to that creature's controller",
			Effect: func(g *game.Game, item *game.StackItem) error {
				if item.Trigger == nil || item.Trigger.Event.CardID == uuid.Nil {
					return nil
				}
				id := item.Trigger.Event.CardID
				var controller uuid.UUID
				if c, ok := g.LookupCardForEffect(id); ok && effects.OnBattlefield(g, id) {
					controller = c.Controller
				} else if info, ok := g.LastKnownPermanentForEffect(id); ok {
					controller = info.Controller
				}
				if controller == uuid.Nil {
					return nil
				}
				return g.DealDamageToPlayerForEffect(item.SourceCardID, controller, 3)
			},
		}},
	})
}
