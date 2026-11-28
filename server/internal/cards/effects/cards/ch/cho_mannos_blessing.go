package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cho-Manno's Blessing — Enchantment — Aura {W}{W} (Mercadian Masques, S82 OLD batch 686, #5619):
//
//	"Flash
//	Enchant creature
//	As this Aura enters, choose a color.
//	Enchanted creature has protection from the chosen color. This effect doesn't remove this Aura."
//
// Ward of Lights' shape with flash: the colour is chosen as the Aura enters and the protection is granted without removing this Aura (CR 702.16n).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fe9f2a67-b8af-48fd-83cd-d15168d2e64f",
		Name:            "Cho-Manno's Blessing",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         effects.EnchantCreature(),
		Purpose:         game.Purpose{Aura: game.AuraBenefit},
		AsEnters:        effects.ChooseColorAsEnters(game.ColorForProtection, "Cho-Manno's Blessing"),
		Static:          []game.StaticAbility{effects.ProtectionFromChosenColorSparingThisAura()},
	})
}
