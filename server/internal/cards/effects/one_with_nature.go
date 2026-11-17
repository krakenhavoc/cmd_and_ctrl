package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// One with Nature — Enchantment — Aura {G} (Scourge):
//
//	"Enchant creature
//	 Whenever enchanted creature deals combat damage to a player, you may
//	 search your library for a basic land card, put that card onto the
//	 battlefield tapped, then shuffle."
//
// Destructive Urge's watcher (the enchanted creature's combat damage to a
// player) as a "you may" asked when the trigger resolves (CR 603.5). "You"
// is the Aura's controller, whoever controls the creature, and the land
// enters under that player's control, tapped.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e48d3de9-a5e0-4a69-8377-9e3ad5fc5b77",
		Name:         "One with Nature",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Triggered: []game.TriggeredAbility{
			Optional(
				On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return attachedCreatureDealtCombatDamageToPlayer(ev, source, g)
				}, "One with Nature — search for a basic land, onto the battlefield tapped", func(g *game.Game, item *game.StackItem) error {
					return SearchLibrary{
						Player:        item.Controller,
						Predicate:     IsBasicLand,
						Dest:          game.ZoneBattlefield,
						Limit:         1,
						TappedOnEntry: true,
						Shuffle:       true,
						Reason:        "One with Nature — a basic land card, onto the battlefield tapped",
					}.Apply(NewContext(g, item))
				}),
				"One with Nature — search your library for a basic land card and put it onto the battlefield tapped?"),
		},
	})
}
