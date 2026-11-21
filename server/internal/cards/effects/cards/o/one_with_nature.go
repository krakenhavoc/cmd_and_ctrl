package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "e48d3de9-a5e0-4a69-8377-9e3ad5fc5b77",
		Name:         "One with Nature",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Triggered: []game.TriggeredAbility{
			effects.Optional(
				effects.On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return effects.AttachedCreatureDealtCombatDamageToPlayer(ev, source, g)
				}, "One with Nature — search for a basic land, onto the battlefield tapped", func(g *game.Game, item *game.StackItem) error {
					return effects.SearchLibrary{
						Player:        item.Controller,
						Predicate:     effects.IsBasicLand,
						Dest:          game.ZoneBattlefield,
						Limit:         1,
						TappedOnEntry: true,
						Shuffle:       true,
						Reason:        "One with Nature — a basic land card, onto the battlefield tapped",
					}.Apply(effects.NewContext(g, item))
				}),
				"One with Nature — search your library for a basic land card and put it onto the battlefield tapped?"),
		},
	})
}
