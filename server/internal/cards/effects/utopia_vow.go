package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "917652c7-ac12-4cd7-ab2d-f2260a620d02",
		Name:         "Utopia Vow",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Grants:       []AbilityGrant{AnyColorManaGrant(utopiaVowGrant)},
		Static: []game.StaticAbility{
			RestrictAttached(game.CantAttackOrBlock),
			Labelled(GrantAbilitiesToAttached(utopiaVowGrant), "Enchanted creature has \"{T}: Add one mana of any color.\""),
		},
	})
}
