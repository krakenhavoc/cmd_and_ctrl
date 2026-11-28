package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Urborg Shambler — Creature — Horror {2}{B}{B}, 4/3 (Invasion, S82 OLD
// #5616):
//
//	"Other black creatures get -1/-1."
//
// A layer 7c static over every other black creature on the battlefield, whoever
// controls it (CR 604.2). Black is a word of the Shambler's text (CR 612.2),
// read through the layers (CR 105.2): a creature that is black through a
// colour-setting effect is shrunk, one that stopped being black is not, and
// the Shambler never shrinks itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7947a210-32d0-4c99-82b4-4b5969447832",
		Name:         "Urborg Shambler",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.OldB652OtherCreaturesOfColorGet("B", -1, -1, "Other black creatures get -1/-1."),
		},
	})
}
