package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Power Taint — Enchantment — Aura {1}{U} (Urza's Saga, #5624):
//
//	"Enchant enchantment
//	 At the beginning of the upkeep of enchanted enchantment's controller, that
//	 player loses 2 life unless they pay {2}.
//	 Cycling {2}"
//
// The player asked is the one whose upkeep it is, which is the enchanted
// enchantment's controller; the question halts the table like any upkeep
// payment (UpkeepPayUnless).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "585a1c82-8a6c-4770-be9c-3a4916b74a9f",
		Name:         "Power Taint",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantPermanent(effects.Enchantment()),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginUpkeep, effects.UpkeepOfEnchantedCreaturesController,
				"Power Taint — that player loses 2 life unless they pay {2}",
				effects.S82OldB725LoseLifeUnlessPaying("Power Taint", "{2}", effects.S82OldB725PowerTaintDecline)),
		},
		Activated: []effects.ActivatedAbility{effects.Cycling("{2}")},
	})
}
