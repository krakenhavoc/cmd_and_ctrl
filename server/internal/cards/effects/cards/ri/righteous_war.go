package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Righteous War — Enchantment {1}{W}{B} (Visions, S82 OLD, #5631):
//
//	"White creatures you control have protection from black.
//	 Black creatures you control have protection from white."
//
// Two Layer-6 keyword grants over the controller's creatures of a colour.
// Colour is read through the layered view, so a creature that is both white
// and black gets both protections.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "acaab3fe-8b12-4f11-9714-7ca278674975",
		Name:         "Righteous War",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.S82OldB788CreaturesYouControlOfColorHaveProtection("W", "protection from black", "White creatures you control have protection from black."),
			effects.S82OldB788CreaturesYouControlOfColorHaveProtection("B", "protection from white", "Black creatures you control have protection from white."),
		},
	})
}
