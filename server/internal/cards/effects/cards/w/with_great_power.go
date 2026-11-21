package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// With Great Power . . . — Enchantment — Aura {3}{W}:
//
//	"Enchant creature you control
//	 Enchanted creature gets +2/+2 for each Aura and Equipment attached to
//	 it.
//	 All damage that would be dealt to you is dealt to enchanted creature
//	 instead."
//
// ADR 0108 §9 decision 4 (#1905): Pariah's static redirection, beside
// Mantle of the Ancients' count of the Auras and Equipment on the host,
// this one included, whoever controls them.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dfedb968-f27c-4117-aff6-da707dd43e82",
		Name:         "With Great Power . . .",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(effects.YouControl()),
		Static:       []game.StaticAbility{effects.PumpAttachedPer(2, 2, effects.AurasAndEquipmentOnTheHost)},
		Replacements: []game.ReplacementEffect{
			effects.RedirectYourDamageToAttached("With Great Power . . . — damage to you is dealt to enchanted creature instead"),
		},
	})
}
