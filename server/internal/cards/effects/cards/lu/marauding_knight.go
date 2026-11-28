package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Marauding Knight — Creature — Phyrexian Zombie Knight {2}{B}{B}, 2/2 (Invasion, S82 OLD
// #5616):
//
//	"Protection from white
//	 This creature gets +1/+1 for each Plains your opponents control."
//
// Protection from white is printed (CR 702.16). The bonus is a layer 7c static
// that counts the Plainss every opponent of its controller controls, read through
// the layers as they stand (CR 604.2, 613.4c): it grows and shrinks as lands
// come and go, and your own Plainss never count. Plains is a word of the
// Knight's text (CR 612.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8745b061-d5c6-46c1-a3fc-4d71d717f19e",
		Name:            "Marauding Knight",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"protection from white"},
		Static: []game.StaticAbility{
			effects.OldB652PumpPerOpponentsPermanent("Plains", 1, 1, "This creature gets +1/+1 for each Plains your opponents control."),
		},
	})
}
