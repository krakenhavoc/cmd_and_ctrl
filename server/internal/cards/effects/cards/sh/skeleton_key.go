package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skeleton Key — Artifact — Equipment {1} (Shadows over Innistrad, #5543):
//
//	"Equipped creature has skulk. (It can't be blocked by creatures with
//	 greater power.)
//	 Whenever equipped creature deals combat damage to a player, you may draw
//	 a card. If you do, discard a card.
//	 Equip {2}"
//
// Skulk is a layer-6 grant. The loot is the Swords' trigger (combat damage
// dealt by the equipped creature to a player) with Sage of the Falls's loot: the
// discard happens only if the draw was taken (CR 608.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "328834f6-1f49-417a-a2d6-d430b8dca177",
		Name:         "Skeleton Key",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.GrantToAttached("skulk")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.AttachedCreatureDealtCombatDamageToPlayer(ev, source, g)
			},
				"Skeleton Key — you may draw a card; if you do, discard a card",
				effects.OldB223MayLoot("Skeleton Key — draw a card, then discard a card?")),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
