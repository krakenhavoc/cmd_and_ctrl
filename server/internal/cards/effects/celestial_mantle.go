package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Celestial Mantle — Enchantment — Aura {3}{W}{W}{W} (Jumpstart Selesnya
// deck request, #4893):
//
//	"Enchant creature
//	 Enchanted creature gets +3/+3.
//	 Whenever enchanted creature deals combat damage to a player, double
//	 its controller's life total."
//
// The pump is a layer 7c static on the host. The trigger watches combat
// damage the enchanted creature deals to a player (the equipment-trigger
// condition the Swords use) and doubles the life total of the creature's
// controller, not the Aura's: an Aura you put on an opponent's creature
// doubles their life. "Its controller" is read as the trigger resolves, or
// from the creature's last-known information once it has left (CR 608.2h).
// Doubling is gaining life equal to the total (CR 701.10e), so a "can't
// gain life" stops it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7417bc73-a75a-4c3a-88d2-17b43e225c2c",
		Name:         "Celestial Mantle",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Static:       []game.StaticAbility{PumpAttached(3, 3)},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return attachedCreatureDealtCombatDamageToPlayer(ev, source, g)
			},
			Key: "Celestial Mantle — double its controller's life total",
			Effect: func(g *game.Game, item *game.StackItem) error {
				if item.Trigger == nil {
					return nil
				}
				return doubleLifeTotal(g, item.SourceCardID, enchantedCreaturesControllerNow(g, item.Trigger.Event))
			},
		}},
	})
}
