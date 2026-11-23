package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chains of Custody — Enchantment — Aura {2}{W} (Secrets of Strixhaven
// Commander, #4834):
//
//	"Enchant creature you control
//	 When this Aura enters, exile target nonland permanent an opponent
//	 controls until this Aura leaves the battlefield.
//	 Enchanted creature has ward {2}."
//
// Banishing Light on an Aura. The enters trigger and its return are
// exileUntilThisLeavesTriggers, the rows Banishing Light and Web Up
// share, so the exiled card comes back when the Aura leaves for any
// reason, including the state-based action when the enchanted creature
// goes. The ward is WardAttached: ward is a triggered ability with a
// cost (CR 702.21a), so the grant is the Aura carrying a ward trigger
// that watches the permanent it is attached to.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5acff7a3-bc84-4251-bd37-80231396ac38",
		Name:         "Chains of Custody",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(effects.YouControl()),
		Triggered: append(effects.ExileUntilThisLeavesTriggers("Chains of Custody"),
			effects.WardAttached(effects.WardMana("{2}"), "Chains of Custody — ward {2}")),
	})
}
