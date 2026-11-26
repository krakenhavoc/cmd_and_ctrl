package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hall of Triumph — Legendary Artifact {3} (Journey into Nyx, S82 OLD, #5558):
//
//	"As Hall of Triumph enters, choose a color.
//	 Creatures you control of the chosen color get +1/+1."
//
// Heraldic Banner's chosen-colour anthem at +1/+1 without the mana ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "716632dd-2947-4641-a1db-50e0a2b0b4e8",
		Name:         "Hall of Triumph",
		Completeness: effects.CompletenessFull,
		AsEnters:     effects.ChooseColorAsEnters(game.ColorForBenefit, "Hall of Triumph"),
		Static:       []game.StaticAbility{effects.ChosenColorAnthem(1, 1)},
	})
}
