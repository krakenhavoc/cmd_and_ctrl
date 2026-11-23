package cardsd

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Destructive Urge — Enchantment — Aura {1}{R}{R} (Urza's Saga):
//
//	"Enchant creature
//	 Whenever enchanted creature deals combat damage to a player, that
//	 player sacrifices a land of their choice."
//
// The Swords' trigger condition (attachedCreatureDealtCombatDamageToPlayer)
// on an Aura, with Mana Vortex's "that player sacrifices a land": the
// damaged player is the event's target, and they choose which land
// goes. A player with no land sacrifices nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5ab62b09-1135-4d22-866b-3a0d64228bf9",
		Name:         "Destructive Urge",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.AttachedCreatureDealtCombatDamageToPlayer(ev, source, g)
			}, "Destructive Urge — that player sacrifices a land", destructiveUrgeSacrifice),
		},
	})
}

func destructiveUrgeSacrifice(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil || item.Trigger.Event.Target == uuid.Nil {
		return nil
	}
	g.PlayerSacrificesForEffect(item.SourceCardID, item.Trigger.Event.Target, effects.SacrificeSpec("a land", effects.Land()),
		"Destructive Urge — sacrifice a land")
	return nil
}
