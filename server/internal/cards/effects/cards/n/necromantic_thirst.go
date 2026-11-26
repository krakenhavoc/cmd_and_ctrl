package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Necromantic Thirst — Enchantment — Aura {2}{B}{B} (Conspiracy, S82 OLD
// batch 362, #5557):
//
//	"Enchant creature
//	 Whenever enchanted creature deals combat damage to a player, you may
//	 return target creature card from your graveyard to your hand."
//
// The trigger is the Aura's own, matched through the attachment (One with
// Nature's watcher), and targets a creature card in the Aura controller's
// graveyard (Karai's shape). "You may" is asked as the trigger resolves
// (CR 603.5); "you" is the Aura's controller, whoever controls the creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "494895db-3da1-4a19-b7a5-ea47817f4e4f",
		Name:         "Necromantic Thirst",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(
				effects.On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return effects.AttachedCreatureDealtCombatDamageToPlayer(ev, source, g)
				}, "Necromantic Thirst — return target creature card from your graveyard to your hand", func(g *game.Game, item *game.StackItem) error {
					return effects.ReturnLegalGraveyardTargetsToHand(effects.NewContext(g, item))
				}),
				effects.TargetCardInGraveyard("target creature card from your graveyard", effects.YouOwn(), effects.Creature())),
				"Necromantic Thirst — return a creature card from your graveyard to your hand?"),
		},
	})
}
