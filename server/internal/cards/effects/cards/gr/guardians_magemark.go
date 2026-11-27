package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Guardian's Magemark — Enchantment — Aura {2}{W} (Guildpact):
//
//	"Flash
//	 Enchant creature
//	 Creatures you control that are enchanted get +1/+1."
//
// Beastmaster's Magemark's anthem: a layer 7c modify on every creature the
// Aura's controller controls that has an Aura attached, whoever controls
// that Aura, so the creature it enchants is one of them.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "95348ba4-038d-4f86-8be0-eb1a68a1f231",
		Name:            "Guardian's Magemark",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         effects.EnchantCreature(),
		Purpose:         game.Purpose{Aura: game.AuraBenefit},
		Static:          []game.StaticAbility{effects.OldB570MagemarkPump()},
	})
}
