package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fencer's Magemark — Enchantment — Aura {2}{R} (Guildpact):
//
//	"Enchant creature
//	 Creatures you control that are enchanted get +1/+1 and have first
//	 strike."
//
// Beastmaster's Magemark's anthem, with a layer 6 first strike grant to the
// same set of creatures.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5f08f79e-bbbe-4af0-ade6-9ad006a5a0d3",
		Name:         "Fencer's Magemark",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.OldB570MagemarkPump(),
			effects.KeywordGrant(effects.OldB570YourEnchantedCreature, "first strike"),
		},
	})
}
