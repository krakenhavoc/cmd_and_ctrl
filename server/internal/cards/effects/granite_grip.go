package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Granite Grip — Enchantment — Aura {2}{R} (Urza's Legacy, #5623):
//
//	"Enchant creature
//	 Enchanted creature gets +1/+0 for each Mountain you control."
//
// Blanchwood Armor's count with Mountains and +1/+0: the Mountains are
// counted per recompute and are the ones the Aura's controller controls, not
// the enchanted creature's controller. "Mountain" is a word of the card's
// text, so a Magical Hack changes what it counts.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "19e09405-3294-4f98-a18f-60cf0a8a2e4e",
		Name:         "Granite Grip",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			PumpAttachedPer(1, 0, mountainsTheAurasControllerControls),
		},
	})
}
