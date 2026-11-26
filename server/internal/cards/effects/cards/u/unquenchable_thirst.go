package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Unquenchable Thirst — {1}{U} Enchantment — Aura (Hour of Devastation):
//
//	"Enchant creature
//	 When this Aura enters, if you control a Desert or there is a Desert card
//	 in your graveyard, tap enchanted creature.
//	 Enchanted creature doesn't untap during its controller's untap step."
//
// The "if" is an intervening condition (CR 603.4), judged as the Aura enters
// and again as the trigger resolves.
func init() {
	effects.Register(effects.Spec{
		OracleID:              "5f55d41a-ef93-4d4f-ab55-13ea887e7af2",
		Name:                  "Unquenchable Thirst",
		Completeness:          effects.CompletenessFull,
		Targets:               effects.EnchantCreature(),
		Purpose:               game.Purpose{Aura: game.AuraHarm},
		UntapStepRestrictions: []game.UntapStepRestriction{effects.EnchantedDoesntUntap()},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIfYou("you control a Desert or there is a Desert card in your graveyard", effects.OldB239YouHaveADesert,
				effects.WhenThisEnters("Unquenchable Thirst — tap enchanted creature", effects.TapEnchantedCreatureOnEntry)),
		},
	})
}
