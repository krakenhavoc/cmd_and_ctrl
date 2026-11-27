package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Primal Cocoon — Enchantment — Aura {G} (Magic 2011 / Planeswalker Decks,
// #5579):
//
//	"Enchant creature
//	 At the beginning of your upkeep, put a +1/+1 counter on enchanted
//	 creature.
//	 When enchanted creature attacks or blocks, sacrifice this Aura."
//
// Two triggers. The upkeep one is the Aura controller's ("your") and puts the
// counter on whatever the Aura is attached to as it resolves, or does nothing
// if it is attached to nothing (Ephara's Enlightenment's read). The second
// fires once per attack or block declaration of the enchanted creature and
// sacrifices the Aura while it is still the same permanent on the
// battlefield; the counters stay.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "79f575b3-ebd7-4d6a-acb9-27c16bcc7a39",
		Name:         "Primal Cocoon",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Primal Cocoon — put a +1/+1 counter on enchanted creature", effects.PlusOneCounterOnEnchantedCreature),
			effects.S82OldB476EnchantedAttacksOrBlocks("Primal Cocoon — sacrifice this Aura", effects.SacrificeThisIfStillOnBattlefield),
		},
	})
}
