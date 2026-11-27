package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jeweled Torque — Artifact {2} (Mercadian Masques, #5619):
//
//	"As this artifact enters, choose a color.
//	 Whenever a player casts a spell of the chosen color, you may pay {2}.
//	 If you do, you gain 2 life."
//
// Prism Ring's as-enters colour, with any player's cast as the trigger and
// a keyed pay-unless: the Torque's controller is asked for {2} and gains
// the life only if it is paid.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "91f36a03-6031-4834-98ee-20b36cdf5569",
		Name:         "Jeweled Torque",
		Completeness: effects.CompletenessFull,
		AsEnters:     effects.ChooseColorAsEnters(game.ColorForBenefit, "Jeweled Torque"),
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.AnyPlayerCastsASpellOfTheChosenColor,
				"Jeweled Torque — you may pay {2}; if you do, you gain 2 life",
				effects.MayPayTwoToGainTwoLife("Jeweled Torque")),
		},
	})
}
