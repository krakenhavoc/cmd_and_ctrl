package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Soul Bleed — Enchantment — Aura {2}{B} (Magic 2010, #5585):
//
//	"Enchant creature
//	 At the beginning of the upkeep of enchanted creature's controller, that
//	 player loses 1 life."
//
// Stab Wound's upkeep trigger without the -2/-2: the player is the enchanted
// creature's controller at the upkeep (CR 603.2), read off the event.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e4f1acd6-b883-470b-88c2-3989011869d7",
		Name:         "Soul Bleed",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginUpkeep, effects.UpkeepOfEnchantedCreaturesController,
				"Soul Bleed — that player loses 1 life",
				func(g *game.Game, item *game.StackItem) error {
					return g.ChangePlayerLifeForEffect(item.SourceCardID, item.Trigger.Event.Actor, -1)
				}),
		},
	})
}
