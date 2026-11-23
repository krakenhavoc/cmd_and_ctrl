package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Utopia Vow — Enchantment — Aura {1}{G} (Planar Chaos, #5595):
//
//	"Enchant creature
//	 Enchanted creature can't attack or block.
//	 Enchanted creature has "{T}: Add one mana of any color.""
//
// Pacifism's restriction (RestrictAttached) and a granted mana ability
// (ADR 0093) on the enchanted creature, AnyColorManaGrant through
// GrantAbilitiesToAttached. Both end the moment the Aura leaves; the
// creature's controller taps it for the mana.
//
// No simplification.
const utopiaVowGrant = "utopia-vow/any-color"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "917652c7-ac12-4cd7-ab2d-f2260a620d02",
		Name:         "Utopia Vow",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Grants:       []effects.AbilityGrant{effects.AnyColorManaGrant(utopiaVowGrant)},
		Static: []game.StaticAbility{
			effects.RestrictAttached(game.CantAttackOrBlock),
			effects.Labelled(effects.GrantAbilitiesToAttached(utopiaVowGrant), "Enchanted creature has \"{T}: Add one mana of any color.\""),
		},
	})
}
