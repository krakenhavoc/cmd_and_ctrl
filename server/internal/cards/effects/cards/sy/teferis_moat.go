package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Teferi's Moat — Enchantment {3}{W}{U} (Invasion):
//
//	"As this enchantment enters, choose a color.
//	 Creatures of the chosen color without flying can't attack you."
//
// The colour is chosen as it enters (CR 614.12). The bar is Storm,
// Windrider's shape (CR 508.1c): the player alone, so a matching creature may
// still attack a planeswalker its controller's opponent controls, with the
// colour read from this enchantment and the flying judged at declaration.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dafd319e-1d4a-4e08-9167-a82ad8670f6f",
		Name:         "Teferi's Moat",
		Completeness: effects.CompletenessFull,
		AsEnters:     effects.ChooseColorAsEnters(game.ColorForHarm, "Teferi's Moat"),
		Static: []game.StaticAbility{
			effects.OldB666CreaturesOfTheChosenColorWithoutFlyingCantAttackYou("Creatures of the chosen color without flying can't attack you."),
		},
	})
}
