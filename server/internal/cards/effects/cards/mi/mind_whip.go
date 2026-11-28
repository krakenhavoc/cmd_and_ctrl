package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mind Whip — Enchantment — Aura {2}{B}{B} (Ice Age, S82 OLD #5636):
//
//	"Enchant creature
//	 At the beginning of the upkeep of enchanted creature's controller, that
//	 player may pay {3}. If they don't, this Aura deals 2 damage to that player
//	 and you tap that creature."
//
// Slow Motion's upkeep pay-unless; declining (or a payment the player cannot
// fund) deals the damage and taps the creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e26f7a58-18ae-432b-bc3f-c410512eb6bf",
		Name:         "Mind Whip",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginUpkeep, effects.UpkeepOfEnchantedCreaturesController,
				"Mind Whip — that player may pay {3}; if they don't, 2 damage to them and tap the creature", effects.MindWhipB830),
		},
	})
}
