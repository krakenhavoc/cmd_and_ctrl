package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Maddening Wind — Enchantment — Aura {2}{G} (Ice Age, S82 OLD batch 839, #5636):
//
//	"Enchant creature
//	 Cumulative upkeep {G}
//	 At the beginning of the upkeep of enchanted creature's controller, this Aura
//	 deals 2 damage to that player."
//
// Soul Bleed's upkeep trigger with damage: the player is the enchanted creature's
// controller at the upkeep (CR 603.2), read off the event. The damage is the
// Aura's, dealt by its last-known information if it has left (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c9798915-3f9e-466b-bbd6-f4e126bdb67d",
		Name:         "Maddening Wind",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Triggered: []game.TriggeredAbility{
			effects.CumulativeUpkeep("Maddening Wind — cumulative upkeep {G}", "{G}"),
			effects.On(game.EventBeginUpkeep, effects.UpkeepOfEnchantedCreaturesController,
				"Maddening Wind — 2 damage to that player",
				func(g *game.Game, item *game.StackItem) error {
					return effects.DealDamage{Source: item.SourceCardID, Target: item.Trigger.Event.Actor, Amount: 2}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
