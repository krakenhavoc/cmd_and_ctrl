package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Multani's Harmony — Enchantment — Aura {G} (Planeshift, #5615):
//
//	"Enchant creature
//	 Enchanted creature has "{T}: Add one mana of any color.""
//
// Utopia Vow's granted mana ability (AnyColorManaGrant through
// GrantAbilitiesToAttached): the creature's controller taps it, it needs the
// creature to be untapped and free of summoning sickness (CR 302.6), and it is
// lost with the Aura.
//
// No simplification.
func init() {
	const grant = "multanis-harmony/any-color"
	effects.Register(effects.Spec{
		OracleID:     "81c223a5-5ff2-43e5-bd0b-25c4f72b2150",
		Name:         "Multani's Harmony",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants:       []effects.AbilityGrant{effects.AnyColorManaGrant(grant)},
		Static: []game.StaticAbility{
			effects.Labelled(effects.GrantAbilitiesToAttached(grant), "Enchanted creature has \"{T}: Add one mana of any color.\""),
		},
	})
}
