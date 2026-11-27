package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Biting Tether — Enchantment — Aura {4}{U} (Shadowmoor, S82 older sets):
//
//	"Enchant creature
//	 You control enchanted creature.
//	 At the beginning of your upkeep, put a -1/-1 counter on enchanted creature."
//
// Mind Control's control effect and Primal Cocoon's upkeep counter, with the
// counter the other kind. The upkeep is the Aura controller's, whoever controls
// the creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1ea20ecb-1919-4be7-b0bb-d79d2dec98fd",
		Name:         "Biting Tether",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Static:       []game.StaticAbility{effects.ControlAttachedBySource()},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Biting Tether — put a -1/-1 counter on enchanted creature", effects.MinusOneCounterOnEnchantedCreature),
		},
	})
}
