package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Xenograft — Enchantment {4}{U} (New Phyrexia, S82 OLD batch 463, #5575):
//
//	"As this enchantment enters, choose a creature type.
//	 Each creature you control is the chosen type in addition to its other
//	 types."
//
// Arcane Adaptation's as-enters choice and its first layer-4 static
// (creatures you control on the battlefield); the creature-spell and
// off-battlefield statics are that card's alone. Before a type is chosen
// nothing is added.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "76642e7a-3c8f-45b8-9cd7-dac8086bea87",
		Name:         "Xenograft",
		Completeness: effects.CompletenessFull,
		AsEnters:     effects.ChooseCreatureTypeAsEnters("Xenograft — choose a creature type"),
		Static:       effects.ChosenTypeInAdditionStatics()[:1],
	})
}
