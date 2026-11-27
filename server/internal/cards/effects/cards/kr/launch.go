package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Launch — Enchantment — Aura {1}{U} (Urza's Saga, #5624):
//
//	"Enchant creature
//	 Enchanted creature has flying.
//	 When this Aura is put into a graveyard from the battlefield, return it to
//	 its owner's hand."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9e3b373f-05f5-44df-9568-bd1f8039aa54",
		Name:         "Launch",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.GrantToAttached("flying")},
		Triggered:    []game.TriggeredAbility{effects.UlgB1AuraReturnsToHand("Launch")},
	})
}
