package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Prismatic Ward — Enchantment — Aura {1}{W} (Ice Age, S82 OLD #5636):
//
//	"Enchant creature
//	 As this Aura enters, choose a color.
//	 Prevent all damage that would be dealt to enchanted creature by sources of
//	 the chosen color."
//
// Benevolent Blessing's colour choice, with a CR 615 prevention on the host in
// place of protection.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ad5e2cb0-00d7-4dad-b20e-23bd41ada398",
		Name:         "Prismatic Ward",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		AsEnters:     effects.ChooseColorAsEnters(game.ColorForProtection, "Prismatic Ward"),
		Replacements: []game.ReplacementEffect{
			effects.PreventDamageToAttachedFromChosenColorB829("Prismatic Ward — prevent all damage that would be dealt to enchanted creature by sources of the chosen color"),
		},
	})
}
