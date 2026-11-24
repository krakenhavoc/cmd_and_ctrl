package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Benevolent Blessing — Enchantment — Aura {1}{W} (Commander Legends):
//
//	"Flash
//	 Enchant creature
//	 As this Aura enters, choose a color.
//	 Enchanted creature has protection from the chosen color. This effect
//	 doesn't remove Auras and Equipment you control that are already attached
//	 to it."
//
// Flickering Ward's as-enters colour choice and protection static, with CR
// 702.16p's exemption in place of 702.16n's: the protection instance it mints
// carries the Aura's controller (game.ProtectionSparingAttachedOf), so the CR
// 704.5m check passes over the Auras and Equipment that player controls
// (this Aura among them) and over nothing else; another instance of
// protection from the same colour, or an attachment controlled by anyone
// else, is removed as normal. Protection already stops anything with the
// colour from becoming attached, so "already attached" needs no further
// record. Before a colour is chosen the creature has no protection.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9472c96a-2827-4af6-a6a1-b6749a5e9069",
		Name:            "Benevolent Blessing",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         effects.EnchantCreature(),
		Purpose:         game.Purpose{Aura: game.AuraBenefit},
		AsEnters:        effects.ChooseColorAsEnters(game.ColorForProtection, "Benevolent Blessing"),
		Static: []game.StaticAbility{
			effects.ProtectionFromChosenColorSparingAttached(),
		},
	})
}
