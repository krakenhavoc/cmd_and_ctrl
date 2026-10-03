package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "dfedb968-f27c-4117-aff6-da707dd43e82",
		Name:         "With Great Power . . .",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(YouControl()),
		Static:       []game.StaticAbility{PumpAttachedPer(2, 2, aurasAndEquipmentOnTheHost)},
		Replacements: []game.ReplacementEffect{
			redirectYourDamageToAttached("With Great Power . . . — damage to you is dealt to enchanted creature instead"),
		},
	})
}
